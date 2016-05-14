package main

import (
	"archive/tar"
	"bytes"

	docker "github.com/fsouza/go-dockerclient"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

func main() {
	client, _ := docker.NewClientFromEnv()
	t := time.Now()
	inputbuf, outputbuf := bytes.NewBuffer(nil), bytes.NewBuffer(nil)
	files := getFilesForTar()

	tw := tar.NewWriter(inputbuf)
	for _, file := range files {
		hdr := &tar.Header{
			Name:       file.Name,
			Size:       int64(len(file.Body)),
			ModTime:    t,
			AccessTime: t,
			ChangeTime: t,
		}
		if err := tw.WriteHeader(hdr); err != nil {
			log.Fatalln(err)
		}
		if _, err := tw.Write([]byte(file.Body)); err != nil {
			log.Fatalln(err)
		}
	}

	if err := tw.Close(); err != nil {
		log.Fatalln(err)
	}

	opts := docker.BuildImageOptions{
		Name:                "repl",
		NoCache:             true,
		SuppressOutput:      false,
		RmTmpContainer:      true,
		ForceRmTmpContainer: true,
		InputStream:         inputbuf,
		OutputStream:        outputbuf,
	}

	if err := client.BuildImage(opts); err != nil {
		log.Fatalln(err)
	}

	log.Println(outputbuf)
}

func getFilesForTar() []struct{ Name, Body string } {
	home := os.Getenv("HOME")
	projectFile, err := ioutil.ReadFile("project.clj")
	if err != nil {
		log.Fatalln(err)
	}

	profileFile, err := ioutil.ReadFile(filepath.Join(home, ".lein/profiles.clj"))
	if err != nil {
		log.Fatalln(err)
	}

	var files = []struct {
		Name, Body string
	}{
		{"project.clj", string(projectFile)},
		{"profiles.clj", string(profileFile)},
		{"Dockerfile", templ},
	}

	return files

}

const templ = `FROM clojure:lein-2.6.1
               COPY profiles.clj .lein/profiles.clj
               COPY project.clj project.clj
               EXPOSE 7888
               CMD ["lein", "repl", ":headless", ":host", "0.0.0.0", ":port", "7888"]`

func genDockerfile() string {
	return templ
}
