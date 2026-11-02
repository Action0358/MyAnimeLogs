package main

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHTTPServer_Timeouts(t *testing.T) {
	srv := newHTTPServer("127.0.0.1:0", http.NewServeMux())

	assert.Equal(t, serverReadHeaderTimeout, srv.ReadHeaderTimeout)
	assert.Equal(t, serverReadTimeout, srv.ReadTimeout)
	assert.Equal(t, serverWriteTimeout, srv.WriteTimeout)
	assert.Equal(t, serverIdleTimeout, srv.IdleTimeout)
}
