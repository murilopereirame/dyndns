package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeUrl(t *testing.T) {
	updater := Client{
		BaseUrl: "https://foo.bar",
	}

	url := updater.MakeUrl("preferences")

	assert.Equal(t, "https://foo.bar/preferences", url)
}

func TestMakeEmptyUrl(t *testing.T) {
	updater := Client{
		BaseUrl: "https://foo.bar",
	}

	url := updater.MakeUrl("")

	assert.Equal(t, "https://foo.bar", url)
}
