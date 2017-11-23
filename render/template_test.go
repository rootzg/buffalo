package render_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/packr"
	"github.com/stretchr/testify/require"
)

func Test_Template(t *testing.T) {
	r := require.New(t)

	tmpFile, err := ioutil.TempFile("", "test")
	r.NoError(err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write([]byte("<%= name %>"))
	r.NoError(err)

	type ji func(string, ...string) render.Renderer

	table := []ji{
		render.New(render.Options{}).Template,
	}

	for _, j := range table {
		re := j("foo/bar", tmpFile.Name())
		r.Equal("foo/bar", re.ContentType())
		bb := &bytes.Buffer{}
		err = re.Render(bb, render.Data{"name": "Mark"})
		r.NoError(err)
		r.Equal("Mark", strings.TrimSpace(bb.String()))
	}
}

func Test_Template_Partial(t *testing.T) {
	r := require.New(t)

	tPath, err := ioutil.TempDir("", "")
	r.NoError(err)
	defer os.Remove(tPath)

	partFile, err := os.Create(filepath.Join(tPath, "_foo.html"))
	r.NoError(err)

	_, err = partFile.Write([]byte("Foo > <%= name %>"))
	r.NoError(err)

	tmpFile, err := os.Create(filepath.Join(tPath, "index.html"))
	r.NoError(err)

	_, err = tmpFile.Write([]byte(`<%= partial("foo.html") %>`))
	r.NoError(err)

	type ji func(string, ...string) render.Renderer

	j := render.New(render.Options{
		TemplatesBox: packr.NewBox(tPath),
	}).Template

	re := j("foo/bar", "index.html")
	r.Equal("foo/bar", re.ContentType())
	bb := &bytes.Buffer{}
	err = re.Render(bb, render.Data{"name": "Mark"})
	r.NoError(err)
	r.Equal("Foo > Mark", strings.TrimSpace(bb.String()))
}

func Test_Template_Partial_WithoutExtension(t *testing.T) {
	r := require.New(t)

	tPath, err := ioutil.TempDir("", "")
	r.NoError(err)
	defer os.Remove(tPath)

	partFile, err := os.Create(filepath.Join(tPath, "_foo.html"))
	r.NoError(err)

	_, err = partFile.Write([]byte("Foo > <%= name %>"))
	r.NoError(err)

	tmpFile, err := os.Create(filepath.Join(tPath, "index.html"))
	r.NoError(err)

	_, err = tmpFile.Write([]byte(`<%= partial("foo") %>`))
	r.NoError(err)

	type ji func(string, ...string) render.Renderer

	j := render.New(render.Options{
		TemplatesBox: packr.NewBox(tPath),
	}).HTML

	re := j("index.html")
	bb := &bytes.Buffer{}
	err = re.Render(bb, render.Data{"name": "Mark"})
	r.NoError(err)
	r.Equal("Foo > Mark", strings.TrimSpace(bb.String()))
}

func Test_AssetPath(t *testing.T) {
	r := require.New(t)

	cases := map[string]string{
		"something.txt":         "/assets/something.txt",
		"images/something.png":  "/assets/images/something.png",
		"/images/something.png": "/assets/images/something.png",
		"application.css":       "/assets/application.aabbc123.css",
	}

	tdir, err := ioutil.TempDir("", "test")
	if err != nil {
		r.Fail("Could not set the Temp dir")
	}

	re := render.New(render.Options{
		AssetsBox: packr.NewBox(tdir),
	}).Template

	ioutil.WriteFile(filepath.Join(tdir, "manifest.json"), []byte(`{
		"application.css": "application.aabbc123.css"
	}`), 0644)

	for original, expected := range cases {

		tmpFile, err := os.Create(filepath.Join(tdir, "test.html"))
		r.NoError(err)

		_, err = tmpFile.Write([]byte("<%= assetPath(\"" + original + "\") %>"))
		r.NoError(err)

		result := re("text/html", tmpFile.Name())

		bb := &bytes.Buffer{}
		err = result.Render(bb, render.Data{})
		r.NoError(err)
		r.Equal(expected, strings.TrimSpace(bb.String()))

		os.Remove(tmpFile.Name())
	}
}

func Test_AssetPathNoManifest(t *testing.T) {
	r := require.New(t)

	cases := map[string]string{
		"something.txt": "/assets/something.txt",
	}

	tdir, err := ioutil.TempDir("", "test")
	if err != nil {
		r.Fail("Could not set the Temp dir")
	}

	re := render.New(render.Options{
		AssetsBox: packr.NewBox(tdir),
	}).Template

	for original, expected := range cases {

		tmpFile, err := os.Create(filepath.Join(tdir, "test.html"))
		r.NoError(err)

		_, err = tmpFile.Write([]byte("<%= assetPath(\"" + original + "\") %>"))
		r.NoError(err)

		result := re("text/html", tmpFile.Name())

		bb := &bytes.Buffer{}
		err = result.Render(bb, render.Data{})
		r.NoError(err)
		r.Equal(expected, strings.TrimSpace(bb.String()))

		os.Remove(tmpFile.Name())
	}
}
func Test_AssetPathManifestCorrupt(t *testing.T) {
	r := require.New(t)

	cases := map[string]string{
		"something.txt": "manifest.json is not correct",
		"other.txt":     "manifest.json is not correct",
	}

	tdir, err := ioutil.TempDir("", "test")
	r.NoError(err)

	ioutil.WriteFile(filepath.Join(tdir, "manifest.json"), []byte(`//shdnn Corrupt!`), 0644)

	re := render.New(render.Options{
		AssetsBox: packr.NewBox(tdir),
	}).Template

	for original, expected := range cases {

		tmpFile, err := os.Create(filepath.Join(tdir, "test.html"))
		r.NoError(err)

		_, err = tmpFile.Write([]byte("<%= assetPath(\"" + original + "\") %>"))
		r.NoError(err)

		result := re("text/html", tmpFile.Name())

		bb := &bytes.Buffer{}
		err = result.Render(bb, render.Data{})
		r.Error(err)
		r.Contains(err.Error(), expected)

		os.Remove(tmpFile.Name())
	}
}

func Test_Suffixes(t *testing.T) {
	r := require.New(t)

	tPath, err := ioutil.TempDir("", "test")
	r.NoError(err)

	tmpFile, err := os.Create(filepath.Join(tPath, "test"))
	r.NoError(err)
	defer os.Remove(tmpFile.Name())

	tmpFileSuffixed, err := os.Create(tmpFile.Name() + "_alt")
	r.NoError(err)
	defer os.Remove(tmpFileSuffixed.Name())

	_, err = tmpFile.Write([]byte("normal"))
	r.NoError(err)

	_, err = tmpFileSuffixed.Write([]byte("suffixed"))
	r.NoError(err)

	type ji func(string, ...string) render.Renderer

	table := []ji{
		render.New(render.Options{
			TemplatesBox: packr.NewBox(tPath),
		}).Template,
	}

	for _, j := range table {
		re := j("foo/bar", filepath.Base(tmpFile.Name()))
		r.Equal("foo/bar", re.ContentType())
		bb1 := &bytes.Buffer{}
		err = re.Render(bb1, render.Data{})
		r.NoError(err)
		r.Equal("normal", strings.TrimSpace(bb1.String()))

		bb2 := &bytes.Buffer{}
		err = re.Render(bb2, render.Data{"templateSuffixes": []string{"alt"}})
		r.NoError(err)
		r.Equal("suffixed", strings.TrimSpace(bb2.String()))
	}
}
