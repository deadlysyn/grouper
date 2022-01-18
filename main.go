package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func main() {
	// default logger and recovery middleware
	r := gin.Default()

	// https://pkg.go.dev/github.com/gin-gonic/gin#section-readme
	h := os.Getenv("TRUSTED_REQUEST_HEADER")
	if len(h) > 0 {
		r.TrustedPlatform = h
	} else {
		p := os.Getenv("TRUSTED_PROXIES")
		var tp []string
		for _, v := range strings.Split(p, " ") {
			tp = append(tp, v)
		}
		if len(tp) > 0 {
			r.SetTrustedProxies(tp)
		}
	}

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "OK",
		})
	})

	v1 := r.Group("/api/v1")
	{
		v1.DELETE("/groups/:groupname/users/:username", deleteGroupUserHandler)
		v1.GET("/groups", getGroupsHandler)
		v1.GET("/groups/:groupname", getGroupUsersHandler)
		v1.GET("/users/:username", getUserHandler)
		v1.POST("/users/:username/groups/:groupname", postUserGroupsHandler)
	}

	// serves on :8080 unless PORT environment variable defined
	err := r.Run()
	if err != nil {
		log.Fatal(err.Error())
	}
}

func handleError(c *gin.Context, status int, msg string) {
	c.JSON(status, gin.H{
		"status":  status,
		"message": msg,
	})
}

// DELETE endpoints

func deleteGroupUserHandler(c *gin.Context) {
	groupname := c.Param("groupname")
	username := c.Param("username")

	req, requester := getRequester(c)
	if len(req.Group) == 0 {
		handleError(c, http.StatusBadRequest, "invalid request")
		return
	}

	// only "admin" group can delete
	if isAdmin(req.KeyID, requester) {
		err := deleteGroupUser(groupname, username, requester)
		if err != nil {
			handleError(c, http.StatusBadGateway, err.Error())
			return
		}

		out, err := getGroupUsers(groupname)
		if err != nil {
			handleError(c, http.StatusBadGateway, err.Error())
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": http.StatusOK,
			"message": gin.H{
				"group":   out.Group,
				"members": out.Users,
			},
		})
	} else {
		handleError(c, http.StatusForbidden, "permission denied")
		return
	}
}

// GET endpoints

func getGroupsHandler(c *gin.Context) {
	out, err := getGroups()
	if err != nil {
		handleError(c, http.StatusBadGateway, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
		"message": gin.H{
			"groups": out.Groups,
		},
	})
}

func getGroupUsersHandler(c *gin.Context) {
	groupname := c.Param("groupname")

	out, err := getGroupUsers(groupname)
	if err != nil {
		handleError(c, http.StatusBadGateway, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
		"message": gin.H{
			"group":   out.Group,
			"members": out.Users,
		},
	})
}

func getUserHandler(c *gin.Context) {
	username := c.Param("username")

	out, err := getUser(username)
	if err != nil {
		handleError(c, http.StatusBadGateway, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
		"message": gin.H{
			"user":   out.User,
			"groups": out.Groups,
		},
	})
}

// POST endpoints

func postUserGroupsHandler(c *gin.Context) {
	username := c.Param("username")
	groupname := c.Param("groupname")
	req, requester := getRequester(c)

	if hasValidKey(req.KeyID, requester) {
		err := updateGroup(groupname, username, requester)
		if err != nil {
			handleError(c, http.StatusBadGateway, err.Error())
			return
		}

		out, err := getUser(username)
		if err != nil {
			handleError(c, http.StatusBadGateway, err.Error())
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"status": http.StatusCreated,
			"message": gin.H{
				"user":   out.User,
				"groups": out.Groups,
			},
		})
	} else {
		handleError(c, http.StatusForbidden, "permission denied")
		return
	}
}
