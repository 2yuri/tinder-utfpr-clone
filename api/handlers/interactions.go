package handlers

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"tinderutf/api/websocket"
	"tinderutf/db/repositories"
	"tinderutf/internal/validate"
)

func FindPeople(c *gin.Context) {
	ctx := context.Background()

	id := c.GetString("userId")
	if id == "" {
		c.AbortWithStatusJSON(500, createDefaultError("invalid userId"))
		return
	}

	user, err := repositories.GetUserById(ctx, id)
	if err != nil {
		c.AbortWithStatusJSON(500, createDefaultError(err.Error()))
		return
	}

	users, err := repositories.FindPeople(ctx, id, user.Options().Distance(), user.Info().Geo(), user.Options().SexPreference(), user.Sex())
	if err != nil {
		c.AbortWithStatusJSON(500, createDefaultError(err.Error()))
		return
	}

	var response []interface{}
	for _, u := range users {
		response = append(response, u.ToJSON())
	}

	if len(response) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	c.JSON(200, createDefaultResponse(response))
}

func ShowLiked(c *gin.Context) {
	ctx := context.Background()

	id := c.GetString("userId")
	if id == "" {
		c.AbortWithStatusJSON(500, createDefaultError("invalid userId"))
		return
	}

	users, err := repositories.ShowLiked(ctx, id)
	if err != nil {
		c.AbortWithStatusJSON(500, createDefaultError(err.Error()))
		return
	}

	var response []interface{}
	for _, u := range users {
		response = append(response, u.ToJSON())
	}

	if len(response) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	c.JSON(200, createDefaultResponse(response))
}

func ShowMatches(c *gin.Context) {
	ctx := context.Background()

	id := c.GetString("userId")
	if id == "" {
		c.AbortWithStatusJSON(500, createDefaultError("invalid userId"))
		return
	}

	users, err := repositories.ShowMatches(ctx, id)
	if err != nil {
		c.AbortWithStatusJSON(500, createDefaultError(err.Error()))
		return
	}

	var response []interface{}
	for _, u := range users {
		response = append(response, u.ToJSON())
	}

	if len(response) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	c.JSON(200, createDefaultResponse(response))
}

func Interact(c *gin.Context) {
	ctx := context.Background()

	id := c.GetString("userId")
	if id == "" {
		c.AbortWithStatusJSON(500, createDefaultError("invalid userId"))
		return
	}

	var req struct {
		Target string `validate:"required" json:"target"`
		Like   bool   `json:"like"`
	}

	if req.Target == id {
		c.AbortWithStatusJSON(500, createDefaultError("invalid target provided."))
		return
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(500, createDefaultError(err.Error()))
		return
	}

	if errors := validate.Struct(&req); len(errors) > 0 {
		c.AbortWithStatusJSON(500, createDefaultError(errors...))
		return
	}

	mutual, err := repositories.CreateInteraction(ctx, id, req.Target, req.Like)
	if err != nil {
		c.AbortWithStatusJSON(500, createDefaultError(err.Error()))
		return
	}

	if mutual && req.Like {
		websocket.Events <- websocket.WSEvent{
			UserID:  req.Target,
			Matched: true,
		}

		websocket.Events <- websocket.WSEvent{
			UserID:  id,
			Matched: true,
		}
	}

	c.Status(200)
}
