package main

import (
    "net/http"

    "github.com/gin-gonic/gin"
)

// album represents data about a record album.
// type album struct {
//     ID     string  `json:"id"`
//     Title  string  `json:"title"`
//     Artist string  `json:"artist"`
//     Price  float64 `json:"price"`
// }

// {
//     "family": image/deb/snap,
//     "name": "name-of-artifact",
//     "version": (version/revision/buildstamp),
//     "test_plan": $application_name,
//     "os": ubuntu/ubuntu-core/ubuntu-flavor,
//     "release": $series,
//     "execution_stage": (pending/proposed/beta),
// }


// albums slice to seed record album data.
var albums = []album{
    {ID: "1", Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
    {ID: "2", Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
    {ID: "3", Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
}


// getAlbums responds with the list of all albums as JSON.
func getAlbums(c *gin.Context) {
    c.IndentedJSON(http.StatusOK, albums)
}

func main() {
    router := gin.Default()
    router.GET("/albums", getAlbums)

    router.Run("localhost:8080")
}


