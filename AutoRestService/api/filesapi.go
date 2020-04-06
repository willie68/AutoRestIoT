package api

import (
	"bufio"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/willie68/AutoRestIoT/dao"
)

//FilesRoutes getting all routes for the config endpoint
func FilesRoutes() *chi.Mux {
	router := chi.NewRouter()
	router.With(RoleCheck([]string{"edit", "read"})).Get("/{bename}/{fileId}", GetFileHandler)
	router.With(RoleCheck([]string{"edit"})).Post("/{bename}/", PostFileEndpoint)
	return router
}

// GetFileHandler get a file
func GetFileHandler(response http.ResponseWriter, request *http.Request) {
	backend := chi.URLParam(request, "bename")
	fileID := chi.URLParam(request, "fileId")
	fmt.Printf("GET: path: %s, be: %s, fileID: %s \n", request.URL.Path, backend, fileID)
	filename, err := dao.GetStorage().GetFilename(backend, fileID)
	if filename == "" {
		render.Render(response, request, ErrNotFound)
		return
	}
	response.Header().Add("Content-disposition", "attachment; filename=\""+filename+"\"")
	err = dao.GetStorage().GetFile(backend, fileID, response)
	if err != nil {
		Msg(response, http.StatusBadRequest, err.Error())
		return
	}
}

//PostFileEndpoint create a new file, return the id
func PostFileEndpoint(response http.ResponseWriter, request *http.Request) {
	backend := chi.URLParam(request, "bename")
	fmt.Printf("POST: path: %s, be: %s\n", request.URL.Path, backend)
	request.ParseForm()
	f, fileHeader, err := request.FormFile("file")
	if err != nil {
		Msg(response, http.StatusBadRequest, err.Error())
		return
	}

	//mimeType := fileHeader.Header.Get("Content-type")
	filename := fileHeader.Filename
	reader := bufio.NewReader(f)

	fileid, err := dao.GetStorage().AddFile(backend, filename, reader)
	if err != nil {
		fmt.Printf("%v\n", err)
	} else {
		fmt.Printf("fileid: %s\n", fileid)
	}

	location := fmt.Sprintf("/api/v1/files/%s/%s", backend, fileid)
	response.Header().Add("Location", location)
	render.Status(request, http.StatusCreated)

	m := make(map[string]interface{})
	m["fileid"] = fileid
	m["filename"] = filename

	render.JSON(response, request, m)
}
