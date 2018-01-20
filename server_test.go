package sshsync

import (
	"github.com/stretchr/testify/assert"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/spf13/afero"
	"testing"
	"net/rpc"
)

func TestServerGetTextFile(t *testing.T) {
	// test just one file because ordering is difficult to compare
	var serverFs = afero.NewMemMapFs()
	// setup
	string1 := "test string 1\nline two"
	// write test data to file
	afero.WriteFile(serverFs, "testFile.txt", []byte(string1), 0644)

	// test
	server := NewServerConfig(serverFs)
	server.buildCache()
	clientConn, serverConn := TwoWayPipe()
	go server.readCommands(serverConn)
	client := rpc.NewClient(clientConn)

	var out string
	err := client.Call(ServerConfig_GetTextFile, "testFile.txt", &out)
	if err != nil {
		t.Fatalf("%s", err)
	}
	assert.Equal(t, out, string1)
	clientConn.Close()
	serverConn.Close()
}

func TestServerGetHashes(t *testing.T) {
	// test just one file because ordering is difficult to compare
	var serverFs = afero.NewMemMapFs()
	// setup
	string1 := "test string 1\nline two"
	// write test data to file
	afero.WriteFile(serverFs, "testFile.txt", []byte(string1), 0644)

	// test
	server := NewServerConfig(serverFs)
	server.buildCache()
	clientConn, serverConn := TwoWayPipe()
	go server.readCommands(serverConn)
	client := rpc.NewClient(clientConn)

	var out ChecksumIndex
	err := client.Call(ServerConfig_GetFileHashes, 0, &out)
	assert.NoError(t, err)
	_, ok := out["testFile.txt"]
	assert.True(t, ok)

	client.Close()
	clientConn.Close()
	serverConn.Close()
}

func TestServerSendTextFile(t *testing.T) {
	// test just one file because ordering is difficult to compare
	var serverFs = afero.NewMemMapFs()
	// setup
	string1 := "test string 1\nline two"
	// write test data to file
	afero.WriteFile(serverFs, "testFile.txt", []byte(string1), 0644)

	// test
	server := NewServerConfig(serverFs)
	server.buildCache()
	clientConn, serverConn := TwoWayPipe()
	go server.readCommands(serverConn)
	client := rpc.NewClient(clientConn)

	// this file should overwrite the existing file
	overwriteFile := TextFile{
		Path:    "testFile.txt",
		Content: "asdfasdfasdf",
	}
	newFile := TextFile{
		Path:    "newpath.cpp",
		Content: "123456789",
	}

	// make sure file has original content
	b, err := afero.ReadFile(serverFs, overwriteFile.Path)
	assert.NoError(t, err)
	assert.Equal(t, string1, string(b))

	// send files
	err = client.Call(ServerConfig_SendTextFile, overwriteFile, nil)
	assert.NoError(t, err)
	err = client.Call(ServerConfig_SendTextFile, newFile, nil)
	assert.NoError(t, err)

	// make sure file is overwritten
	b, err = afero.ReadFile(serverFs, overwriteFile.Path)
	assert.NoError(t, err)
	assert.Equal(t, overwriteFile.Content, string(b))
	// new file is there too
	b, err = afero.ReadFile(serverFs, newFile.Path)
	assert.NoError(t, err)
	assert.Equal(t, newFile.Content, string(b))

	client.Close()
	clientConn.Close()
	serverConn.Close()
}

func TestServerDelta(t *testing.T) {
	var serverFs = afero.NewMemMapFs()
	// setup
	string1 := "test string 1\nline two"
	string2 := "tested string 222\nline 2"
	// write test data to file
	afero.WriteFile(serverFs, "testFile.txt", []byte(string1), 0644)
	// get Delta
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(string1, string2, false)
	delta := dmp.DiffToDelta(diffs)

	// create server
	server := NewServerConfig(serverFs)
	server.buildCache()
	clientConn, serverConn := TwoWayPipe()
	go server.readCommands(serverConn)

	// test call
	client := rpc.NewClient(clientConn)
	err := client.Call(ServerConfig_Delta, TextFileDeltas{
		{
			Path:  "testFile.txt",
			Delta: delta,
		}}, nil)
	assert.NoError(t, err)

	// verify file now contains string2
	fileBytes, err := afero.ReadFile(serverFs, "testFile.txt")
	assert.NoError(t, err)
	assert.Equal(t, string2, string(fileBytes))

	client.Close()
	clientConn.Close()
	serverConn.Close()
}
