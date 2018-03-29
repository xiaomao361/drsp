package main

// A PACS server. Supports C-STORE
//
// Usage: ./server -dir <directory> -port 11111
//
// It starts a DICOM server and serves files under <directory>.

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/grailbio/go-dicom"
	"github.com/grailbio/go-dicom/dicomio"
	"github.com/grailbio/go-dicom/dicomtag"
	"github.com/grailbio/go-netdicom"
	"github.com/grailbio/go-netdicom/dimse"
)

var (
	portFlag   = flag.String("port", "11113", "TCP port to listen to")
	aeFlag     = flag.String("ae", "pi", "AE title of this server")
	outputFlag = flag.String("output", "", `
The directory to store files received by C-STORE.
If empty, use <dir>/dicom, where <dir> is the value of the -dir flag.`)
)

type server struct {
	mu *sync.Mutex

	pathSeq int32
}

func (ss *server) onCStore(transferSyntaxUID string, sopClassUID string, sopInstanceUID string, data []byte) dimse.Status {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.pathSeq++
	path := path.Join(*outputFlag, fmt.Sprintf("image%04d.dcm", ss.pathSeq))
	out, err := os.Create(path)
	if err != nil {
		dirPath := filepath.Dir(path)
		err := os.MkdirAll(dirPath, 0755)
		if err != nil {
			return dimse.Status{Status: dimse.StatusNotAuthorized, ErrorComment: err.Error()}
		}
		out, err = os.Create(path)
		if err != nil {
			log.Printf("%s: create: %v", path, err)
			return dimse.Status{Status: dimse.StatusNotAuthorized, ErrorComment: err.Error()}
		}
	}
	defer func() {
		if out != nil {
			out.Close()
		}
	}()

	e := dicomio.NewEncoderWithTransferSyntax(out, transferSyntaxUID)
	dicom.WriteFileHeader(e,
		[]*dicom.Element{
			dicom.MustNewElement(dicomtag.TransferSyntaxUID, transferSyntaxUID),
			dicom.MustNewElement(dicomtag.MediaStorageSOPClassUID, sopClassUID),
			dicom.MustNewElement(dicomtag.MediaStorageSOPInstanceUID, sopInstanceUID),
		})
	e.WriteBytes(data)
	if err := e.Error(); err != nil {
		log.Printf("%s: write: %v", path, err)
		return dimse.Status{Status: dimse.StatusNotAuthorized, ErrorComment: err.Error()}
	}
	err = out.Close()
	out = nil
	if err != nil {
		log.Printf("%s: close %s", path, err)
		return dimse.Status{Status: dimse.StatusNotAuthorized, ErrorComment: err.Error()}
	}
	log.Printf("C-STORE: Created %v", path)
	return dimse.Success
}

func canonicalizeHostPort(addr string) string {
	if !strings.Contains(addr, ":") {
		return ":" + addr
	}
	return addr
}

func main() {
	flag.Parse()
	port := canonicalizeHostPort(*portFlag)
	if *outputFlag == "" {
		*outputFlag = filepath.Join("dicom")
	}

	ss := server{
		mu: &sync.Mutex{},
	}
	log.Printf("Dciom Server Listening on %s", port)

	params := netdicom.ServiceProviderParams{
		AETitle: *aeFlag,
		CEcho: func(connState netdicom.ConnectionState) dimse.Status {
			log.Printf("Received C-ECHO")
			return dimse.Success
		},
		CStore: func(connState netdicom.ConnectionState, transferSyntaxUID string,
			sopClassUID string,
			sopInstanceUID string,
			data []byte) dimse.Status {
			return ss.onCStore(transferSyntaxUID, sopClassUID, sopInstanceUID, data)
		},
	}
	sp, err := netdicom.NewServiceProvider(params, port)
	if err != nil {
		panic(err)
	}
	sp.Run()
}
