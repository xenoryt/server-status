package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"path"
	"strings"

	"github.com/xenoryt/stream-status/lib/stream"
	"golang.org/x/tools/godoc/vfs"
	"golang.org/x/xerrors"
)

type StreamRequest struct {
	Path string `json:"path"`
	Subs bool   `json:"subs"`
}

func StartServer() {
	fSys, err := fs.Sub(client, "client/dist")
	if err != nil {
		log.Fatalln(err)
	}
	http.Handle("/", http.FileServer(http.FS(fSys)))

	http.HandleFunc("/files/", HandleFilesRequest)
	http.HandleFunc("/stream", HandleStreamRequest)
	http.HandleFunc("/streams", HandleStreamsListRequest)
	http.HandleFunc("/stream-url", HandleStreamUrlRequest)

	logger.Println("Starting server on port 8080...")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalln(err)
	}
}

func toStreamPath(filepath string) string {
	h := md5.Sum([]byte(filepath))
	return path.Join(hex.EncodeToString(h[:]), "out.m3u8")
}

func HandleStreamUrlRequest(w http.ResponseWriter, r *http.Request) {
	filepath := r.URL.Query().Get("path")
	if filepath == "" {
		WriteInvalidRequest(w, fmt.Errorf("Missing `path`"))
		return
	}

	url := fmt.Sprintf("%s/%s", HLSStaticURL, toStreamPath(filepath))
	WriteResponse(w, url, nil)
}

// HandleStreamsListRequest handles requests for old streams
func HandleStreamsListRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// If filepath is specified, return single stream.
		// Else return all streams
		if filepath := r.URL.Query().Get("path"); filepath != "" {
			s := stream.GetStream(filepath)
			if s == nil {
				WriteResponse(w, false, nil)
			}
			WriteResponse(w, !s.Done, nil)
		} else {
			streams := stream.ListStreams()
			streamResponse := make([]map[string]any, len(streams))
			for i, s := range streams {
				streamResponse[i] = map[string]any{
					"url":    fmt.Sprintf("%s/%s", HLSStaticURL, toStreamPath(s.Filepath)),
					"stream": s,
				}
			}
			WriteResponse(w, streamResponse, nil)
		}
	default:
		WriteInvalidRequest(w, fmt.Errorf("Method is not supported"))
	}
}

// HandleStreamRequest handles requests for new stream
func HandleStreamRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodDelete:
		filepath := r.URL.Query().Get("path")
		log.Printf("Stopping stream: %s\n", filepath)
		WriteResponse(w, nil, stream.StopStream(filepath))
	case http.MethodPost:
		var request StreamRequest
		d := json.NewDecoder(r.Body)
		err := d.Decode(&request)
		if err != nil {
			WriteInvalidRequest(w, err)
			return
		}
		filepath := path.Join(*mediaDir, request.Path)
		log.Println("Starting stream on file:", filepath)
		WriteResponse(w, nil, stream.StreamFile(path.Join(*streamOutDir, toStreamPath(request.Path)), filepath, request.Subs))
	default:
		WriteInvalidRequest(w, fmt.Errorf("Method is not supported"))
	}
}

func HandleFilesRequest(w http.ResponseWriter, r *http.Request) {
	tree, err := DirTree(vfs.OS(*mediaDir), "/")
	WriteResponse(w, tree, err)
}

// DirTree returns all contents of a directory in a recursive tree structure.
func DirTree(root vfs.FileSystem, p string) (*FileTreeNode, error) {
	stat, err := root.Stat(p)
	if err != nil {
		return nil, xerrors.Errorf("Failed to stat file %s: %w", p, err)
	}

	children := make([]*FileTreeNode, 0)
	if stat.IsDir() {
		entries, err := root.ReadDir(p)
		if err != nil {
			log.Printf("Failed to read dir '%s': %v", p, err)
		}
		for _, f := range entries {
			childpath := path.Join(p, f.Name())
			// Skip hidden files and text files
			isValid, err := isDirOrVideo(root, childpath)
			if err != nil {
				log.Printf("Failed to check mimetype of %s: %v", childpath, err)
				continue
			}
			if strings.HasPrefix(f.Name(), ".") || !isValid {
				continue
			}
			child, err := DirTree(root, childpath)
			if err != nil {
				log.Printf("Failed to read child %s: %v", f.Name(), err)
				continue
			}
			children = append(children, child)
		}
	}

	return &FileTreeNode{
		Name:  stat.Name(),
		Path:  p,
		IsDir: stat.IsDir(),

		Children: children,
	}, nil
}

func isDirOrVideo(root vfs.FileSystem, filepath string) (bool, error) {
	stat, err := root.Stat(filepath)
	if err != nil {
		return false, xerrors.Errorf("Failed to stat file: %w", err)
	}
	if stat.IsDir() {
		return true, nil
	}
	// mimetype requires reading the file which is too slow.
	//file, err := root.Open(filepath)
	//if err != nil {
	//	return false, xerrors.Errorf("Failed to open file: %w", err)
	//}
	//mtype, err := mimetype.DetectReader(file)
	//if err != nil {
	//	return false, xerrors.Errorf("Failed to check file mimetype: %w", err)
	//}
	//return strings.HasPrefix(mtype.String(), "video/"), nil
	return strings.HasSuffix(stat.Name(), ".mkv") ||
		strings.HasSuffix(stat.Name(), ".mov") ||
		strings.HasSuffix(stat.Name(), ".mp4"), nil
}

func WriteInvalidRequest(w http.ResponseWriter, err error) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	WriteJson(w, NewResponse(nil, err))
}

func WriteResponse(w http.ResponseWriter, data any, err error) {
	status := http.StatusOK
	if err != nil {
		status = http.StatusInternalServerError
		log.Println("ERROR:", err)
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	WriteJson(w, NewResponse(data, err))
}

func WriteJson(w io.Writer, data any) {
	e := json.NewEncoder(w)
	//e.SetEscapeHTML(false)
	e.Encode(data)
}

type Response struct {
	Data  any    `json:"data"`
	Error string `json:"error"`
}

func NewResponse(data any, err error) Response {
	e := ""
	if err != nil {
		e = err.Error()
	}
	return Response{data, e}
}

type FileTreeNode struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	IsDir bool   `json:"isDir"`

	Children []*FileTreeNode `json:"children"`
}
