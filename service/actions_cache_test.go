package service

import (
	"errors"
	"net/url"
	"testing"

	"github.com/actions/gh-actions-cache/internal"
	"github.com/actions/gh-actions-cache/types"
	"github.com/cli/go-gh/pkg/api"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

const VERSION string = "1.0.0"

func TestGetCacheUsage_CorrectRepo(t *testing.T) {
	t.Cleanup(gock.Off)

	gock.New("https://api.github.com").
		Get("/repos/testOrg/testRepo/actions/cache/usage").
		Reply(200).
		JSON(`{
			"full_name": "testOrg/testRepo",
			"active_caches_size_in_bytes": 291205,
			"active_caches_count": 12
		}`)

	repo, err := internal.GetRepo("testOrg/testRepo")
	assert.NoError(t, err)

	artifactCache, err := NewArtifactCache(repo, "list", VERSION)
	assert.NoError(t, err)
	totalCacheSize, err := artifactCache.GetCacheUsage()

	assert.Equal(t, float64(291205), totalCacheSize)
	assert.NoError(t, err)
	assert.True(t, gock.IsDone(), internal.PrintPendingMocks(gock.Pending()))
}

func TestGetCacheUsage_IncorrectRepo(t *testing.T) {
	t.Cleanup(gock.Off)

	gock.New("https://api.github.com").
		Get("/repos/testOrg/wrongRepo/actions/cache/usage").
		Reply(404).
		JSON(`{
			"message": "Not Found",
			"documentation_url": "https://docs.github.com/rest/reference/actions#get-github-actions-cache-usage-for-a-repository"
		}`)

	repo, err := internal.GetRepo("testOrg/wrongRepo")
	assert.NoError(t, err)

	artifactCache, err := NewArtifactCache(repo, "list", VERSION)
	assert.NoError(t, err)
	totalCacheSize, err := artifactCache.GetCacheUsage()
	var httpError api.HTTPError
	errors.As(err, &httpError)

	assert.Error(t, err)
	assert.Equal(t, 404, httpError.StatusCode)
	assert.Equal(t, "Not Found", httpError.Message)
	assert.Equal(t, float64(-1), totalCacheSize)
	assert.True(t, gock.IsDone(), internal.PrintPendingMocks(gock.Pending()))
}

func TestListCaches_Success(t *testing.T) {
	t.Cleanup(gock.Off)

	gock.New("https://api.github.com").
		Get("/repos/testOrg/testRepo/actions/caches").
		Reply(200).
		JSON(`{
			"total_count": 1,
			"actions_caches": [
				{
					"id": 29,
					"ref": "refs/heads/master",
					"key": "Linux-build-cache-node-modules-3fd22dd3a926d576e2562e8b76a5ff157cd3b986f3d44195acfe7efa6bc05919-8",
					"version": "7fcda33c1e1d849a13bcc06f49b9ab64efc01ca9dabe4d7a8d0d387feef4fc88",
					"last_accessed_at": "2022-06-22T20:32:45.550000000Z",
					"created_at": "2022-06-22T20:32:45.550000000Z",
					"size_in_bytes": 2432967
				}]
			}`)

	repo, err := internal.GetRepo("testOrg/testRepo")
	assert.NoError(t, err)

	f := types.ListOptions{BaseOptions: types.BaseOptions{Repo: "testOrg/testRepo"}, Limit: 30}
	queryParams := url.Values{}
	f.GenerateQueryParams(queryParams)

	artifactCache, err := NewArtifactCache(repo, "list", VERSION)
	assert.NoError(t, err)
	listCacheResponse, err := artifactCache.ListCaches(queryParams)

	assert.NoError(t, err)
	assert.NotNil(t, listCacheResponse)
	assert.Equal(t, 1, listCacheResponse.TotalCount)
	assert.Equal(t, 1, len(listCacheResponse.ActionsCaches))
	assert.Equal(t, 29, listCacheResponse.ActionsCaches[0].Id)
	assert.True(t, gock.IsDone(), internal.PrintPendingMocks(gock.Pending()))
}

func TestListCaches_Failure(t *testing.T) {
	t.Cleanup(gock.Off)

	gock.New("https://api.github.com").
		Get("/repos/testOrg/testRepo/actions/caches").
		Reply(404).
		JSON(`{
			"message": "Not Found",
			"documentation_url": "https://docs.github.com/rest/reference/actions#get-github-actions-cache-list-for-a-repository"
		}`)

	repo, err := internal.GetRepo("testOrg/testRepo")
	assert.NoError(t, err)

	f := types.ListOptions{BaseOptions: types.BaseOptions{Repo: "testOrg/testRepo"}, Limit: 30}
	queryParams := url.Values{}
	f.GenerateQueryParams(queryParams)

	artifactCache, err := NewArtifactCache(repo, "list", VERSION)
	assert.NoError(t, err)
	listCacheResponse, err := artifactCache.ListCaches(queryParams)
	var httpError api.HTTPError
	errors.As(err, &httpError)

	assert.Error(t, err)
	assert.Equal(t, 404, httpError.StatusCode)
	assert.Equal(t, "Not Found", httpError.Message)
	assert.Equal(t, types.ListApiResponse{}, listCacheResponse)
	assert.True(t, gock.IsDone(), internal.PrintPendingMocks(gock.Pending()))
}

func TestDeleteCaches_Success(t *testing.T) {
	t.Cleanup(gock.Off)

	gock.New("https://api.github.com").
		Delete("/repos/testOrg/testRepo/actions/caches").
		Reply(200).
		JSON(`{
			"total_count": 1,
			"actions_caches": [
				{
					"id": 29,
					"ref": "refs/heads/master",
					"key": "Linux-build-cache-node-modules-3fd22dd3a926d576e2562e8b76a5ff157cd3b986f3d44195acfe7efa6bc05919-8",
					"version": "7fcda33c1e1d849a13bcc06f49b9ab64efc01ca9dabe4d7a8d0d387feef4fc88",
					"last_accessed_at": "2022-06-22T20:32:45.550000000Z",
					"created_at": "2022-06-22T20:32:45.550000000Z",
					"size_in_bytes": 2432967
				}]
			}`)

	repo, err := internal.GetRepo("testOrg/testRepo")
	assert.NoError(t, err)

	f := types.DeleteOptions{BaseOptions: types.BaseOptions{Repo: "testOrg/testRepo"}}
	queryParams := url.Values{}
	f.GenerateBaseQueryParams(queryParams)

	artifactCache, err := NewArtifactCache(repo, "delete", VERSION)
	assert.NoError(t, err)
	deletedCache, err := artifactCache.DeleteCaches(queryParams)

	assert.NoError(t, err)
	assert.Equal(t, 1, deletedCache)
	assert.True(t, gock.IsDone(), internal.PrintPendingMocks(gock.Pending()))
}

func TestDeleteCaches_Failure(t *testing.T) {
	t.Cleanup(gock.Off)

	gock.New("https://api.github.com").
		Delete("/repos/testOrg/testRepo/actions/caches").
		Reply(404).
		JSON(`{
			"message": "Not Found",
			"documentation_url": "https://docs.github.com/rest/reference/actions#get-github-actions-cache-list-for-a-repository"
		}`)

	repo, err := internal.GetRepo("testOrg/testRepo")
	assert.NoError(t, err)

	f := types.DeleteOptions{BaseOptions: types.BaseOptions{Repo: "testOrg/testRepo"}}
	queryParams := url.Values{}
	f.GenerateBaseQueryParams(queryParams)

	artifactCache, err := NewArtifactCache(repo, "delete", VERSION)
	assert.NoError(t, err)
	deletedCache, err := artifactCache.DeleteCaches(queryParams)

	assert.Error(t, err)
	assert.Equal(t, 0, deletedCache)
	assert.True(t, gock.IsDone(), internal.PrintPendingMocks(gock.Pending()))
}
