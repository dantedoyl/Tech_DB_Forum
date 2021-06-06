package forum

import "github.com/dantedoyl/Tech_DB_Forum/internal/models"

type ForumRepository interface {
	CreateForum(forum *models.Forum) *models.Error
	GetForumInfo(slug string) (*models.Forum, *models.Error)
	CreateThread(thread *models.Thread) *models.Error
	GetForumUsers(slug string, params *models.Params) ([]*models.User, *models.Error)
	GetForumThreads(slug string, params *models.Params) ([]*models.Thread, *models.Error)
	UpdatePostInfo(info *models.PostUpdate) (*models.Post, *models.Error)
	PostInfo(id int, related models.Related) (*models.PostInfo, *models.Error)
	StatusDB() *models.Status
	ClearDB() *models.Error
	CreateUser(user *models.User)([]*models.User, *models.Error)
	GetUserProfile(nickname string) (*models.User, *models.Error)
	UpdateUserProfile(user *models.User) *models.Error
	CreatePosts(posts []*models.Post, slugOrID string) ([]*models.Post, *models.Error)
	GetThreadInfo(slugOrID string) (*models.Thread, *models.Error)
	UpdateThreadInfo(thread *models.Thread) *models.Error
	InsertOrUpdateVote(slugOrID string, vote *models.Vote) (*models.Thread, *models.Error)
	GetThreadPosts(slugOrID string, params *models.Params) ([]*models.Post, *models.Error)
}