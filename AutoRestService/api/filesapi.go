package api

import (
	"bufio"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/willie68/AutoRestIoT/dao"
)

//SchematicsRoutes getting all routes for the config endpoint
func FilesRoutes() *chi.Mux {
	router := chi.NewRouter()
	router.Get("/{bename}/{fileId}", GetFileHandler)
	router.Post("/{bename}/", PostFileEndpoint)
	return router
}

// GetSchematicFileHandler gets a tenant
func GetFileHandler(response http.ResponseWriter, req *http.Request) {
	backend := chi.URLParam(req, "bename")
	fileID := chi.URLParam(req, "fileId")
	fmt.Printf("GET: path: %s, be: %s, fileID: %s \n", req.URL.Path, backend, fileID)
	filename, err := dao.GetStorage().GetFilename(fileID)
	response.Header().Add("Content-disposition", "attachment; filename=\""+filename+"\"")
	err = dao.GetStorage().GetFile(fileID, response)
	if err != nil {
		Msg(response, http.StatusBadRequest, err.Error())
		return
	}
}

//PostFileEndpoint create a new file, return the id
func PostFileEndpoint(response http.ResponseWriter, req *http.Request) {
	backend := chi.URLParam(req, "bename")
	fmt.Printf("POST: path: %s, be: %s\n", req.URL.Path, backend)
	req.ParseForm()
	f, fileHeader, err := req.FormFile("file")
	if err != nil {
		Msg(response, http.StatusBadRequest, err.Error())
		return
	}

	//mimeType := fileHeader.Header.Get("Content-type")
	filename := fileHeader.Filename
	reader := bufio.NewReader(f)

	fileid, err := dao.GetStorage().AddFile(filename, reader)
	if err != nil {
		fmt.Printf("%v\n", err)
	} else {
		fmt.Printf("fileid: %s\n", fileid)
	}

	location := fmt.Sprintf("/api/v1/files/%s", fileid)
	response.Header().Add("Location", location)
	render.Status(req, http.StatusCreated)

	m := make(map[string]interface{})
	m["fileid"] = fileid
	m["filename"] = filename

	render.JSON(response, req, m)
}
