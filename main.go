package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// default logger and recovery middleware
	r := gin.Default()

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "OK",
		})
	})
	r.GET("/groups", getGroupsHandler)
	r.GET("/groups/:groupname", getGroupUsersHandler)
	r.GET("/users/:username", getUserHandler)
	r.POST("/users/:username/groups", postUserGroupsHandler)

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
			// "user":   out.User.User,
			// "groups": out.Groups.Groups,
		},
	})
}

// POST endpoints

func postUserGroupsHandler(c *gin.Context) {
	username := c.Param("username")
	req, requester := getRequester(c)

	if len(req.Group) == 0 {
		handleError(c, http.StatusBadRequest, "invalid request")
		return
	}

	if hasValidKey(req.KeyID, requester) {
		err := updateGroup(req.Group, username, requester)
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
				"user":   out.User.User,
				"groups": out.Groups.Groups,
			},
		})
	} else {
		handleError(c, http.StatusForbidden, "permission denied")
		return
	}
}
