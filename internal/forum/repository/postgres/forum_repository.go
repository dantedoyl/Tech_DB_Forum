package postgres

import (
	"fmt"
	"github.com/dantedoyl/Tech_DB_Forum/internal/models"
	"github.com/jackc/pgx"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ForumRepository struct {
	dbConn *pgx.ConnPool
}

func NewForumRepository( conn *pgx.ConnPool) *ForumRepository{
	return &ForumRepository{dbConn: conn}
}

func (fr ForumRepository) CreateForum(forum *models.Forum) *models.Error{
	var user string
	err := fr.dbConn.QueryRow(`SELECT nickname FROM users WHERE nickname=$1;`, forum.User).Scan(&user)
	if err != nil {
		return &models.Error{Code:    http.StatusNotFound, Message: "Can't find user with nickname"}
	}
	forum.User = user
	err = fr.dbConn.QueryRow(`INSERT INTO forum(slug, author, title) VALUES ($1, $2, $3) RETURNING slug`,
		forum.Slug, forum.User, forum.Title).Scan(&forum.Slug)
	if err != nil {
		if err.(pgx.PgError).Code == "23505"{
			row := fr.dbConn.QueryRow(`SELECT slug, author, title, posts, threads FROM forum
				WHERE slug=$1`, forum.Slug)
			err = row.Scan(&forum.Slug, &forum.User, &forum.Title, &forum.Posts, &forum.Threads)
			return &models.Error{Code: http.StatusConflict}
		} else {
			return &models.Error{Code: http.StatusInternalServerError, Message: err.Error()}
		}
	}

	return nil
}

func (fr ForumRepository)GetForumInfo(slug string) (*models.Forum, *models.Error){
	forum := &models.Forum{}
	row := fr.dbConn.QueryRow(`SELECT slug, author, title, posts, threads FROM forum
				WHERE slug=$1`, slug)
	err := row.Scan(&forum.Slug, &forum.User, &forum.Title, &forum.Posts, &forum.Threads)
	if err != nil {
		return nil, &models.Error{Code: http.StatusNotFound, Message: "Can't find forum" }
	}
	return forum, nil
}

func (fr ForumRepository)CreateThread(thread *models.Thread) *models.Error{
	var forumSlug string
	err := fr.dbConn.QueryRow(`SELECT slug FROM forum
				WHERE slug=$1`, thread.Forum).Scan(&forumSlug)
	if err != nil {
		return &models.Error{Code: http.StatusNotFound, Message: "Can't find forum" }
	}

	var userName string
	row := fr.dbConn.QueryRow(`SELECT nickname FROM users WHERE nickname=$1;`, thread.Author)
	err = row.Scan(&userName)
	if err != nil {
		return &models.Error{Code: http.StatusNotFound, Message: "Can't find user" }
	}
	var slug string
	if thread.Slug == "" {
		slug = thread.Title + thread.Author
	} else {
		slug = thread.Slug
	}
	err = fr.dbConn.QueryRow(	`INSERT INTO thread(title, author, created, forum, message, slug, votes)
							VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, created`, thread.Title, thread.Author, thread.Created,
		thread.Forum,
		thread.Message, slug, thread.Votes).Scan(&thread.ID, &thread.Created)
	if err != nil {
		if err.(pgx.PgError).Code == "23505"{
			row = fr.dbConn.QueryRow(`SELECT id, title, author, forum, message, votes, slug, created FROM thread
							WHERE slug=$1;`, slug)
			err = row.Scan(&thread.ID, &thread.Title, &thread.Author, &thread.Forum, &thread.Message, &thread.Votes,
				&thread.Slug, &thread.Created)
			return &models.Error{Code: http.StatusConflict}
		} else {
			return &models.Error{Code: http.StatusInternalServerError, Message: err.Error()}
		}
	}
	thread.Forum = forumSlug
	return nil
}

func (fr ForumRepository)GetForumUsers(slug string, params *models.Params) ([]*models.User, *models.Error){
	var forumSlug string
	err := fr.dbConn.QueryRow(`SELECT slug FROM forum
				WHERE slug=$1`, slug).Scan(&forumSlug)
	if err != nil {
		return nil, &models.Error{Code: http.StatusNotFound, Message: "Can't find forum" }
	}

	query := `SELECT about, email, fullname, nickname 
				FROM forum_users WHERE slug=$1`
	if params.Desc {
		if params.Since != "" {
			query += fmt.Sprintf(` AND nickname < '%s'`, params.Since)
		}
		query += ` ORDER BY nickname DESC LIMIT NULLIF($2, 0)`
	} else {
		if params.Since != "" {
			query += fmt.Sprintf(` AND nickname > '%s'`, params.Since)
		}
		query += ` ORDER BY nickname LIMIT NULLIF($2, 0)`
	}
	var users []*models.User
	rows, err := fr.dbConn.Query(query, slug, params.Limit)
	if err != nil {
		return nil, &models.Error{Code: http.StatusInternalServerError, Message: err.Error()}
	}

	defer rows.Close()

	for rows.Next() {
		user := &models.User{}
		err = rows.Scan(&user.About, &user.Email, &user.FullName, &user.Nickname)
		if err != nil {
			return nil, &models.Error{Code: http.StatusInternalServerError, Message: err.Error()}
		}
		users = append(users, user)
	}
	return users, nil
}

func (fr ForumRepository)GetForumThreads(slug string, params *models.Params) ([]*models.Thread, *models.Error){
	var forumSlug string
	row := fr.dbConn.QueryRow(`SELECT slug FROM forum
				WHERE slug=$1`, slug)
	err := row.Scan(&forumSlug)
	if err != nil {
		return nil, &models.Error{Code: http.StatusNotFound, Message: "Can't find forum" }
	}

	var threads []*models.Thread
	query := `SELECT id, author, created, forum, message, slug, title, votes FROM thread
		WHERE forum=$1`
	var param []interface{}
	if params.Since != "" {
		if params.Desc {
			query += ` AND created <= $2 ORDER BY created DESC LIMIT $3;`
		} else {
			query += ` AND created >= $2 ORDER BY created ASC LIMIT $3;`
		}
		param = append(param, slug, params.Since, params.Limit)
	} else {
		if params.Desc {
			query += ` ORDER BY created DESC LIMIT $2;`
		} else {
			query += ` ORDER BY created ASC LIMIT $2;`
		}
		param = append(param, slug, params.Limit)
	}

	rows, err := fr.dbConn.Query(query, param...)
	if err != nil {
		return nil, &models.Error{Code: http.StatusInternalServerError, Message: err.Error()}
	}
	defer rows.Close()


	for rows.Next() {
		 thread := &models.Thread{}
		err = rows.Scan(&thread.ID, &thread.Author, &thread.Created, &thread.Forum, &thread.Message,
			&thread.Slug, &thread.Title, &thread.Votes)
		if err != nil {
			return nil, &models.Error{Code: http.StatusInternalServerError, Message: err.Error()}
		}
		threads = append(threads, thread)
	}
	return threads, nil
}

func (fr ForumRepository)UpdatePostInfo(info *models.PostUpdate) (*models.Post, *models.Error){
	var postID int
	row := fr.dbConn.QueryRow(`SELECT id FROM post WHERE id=$1;`, info.ID)
	err := row.Scan(&postID)
	if err != nil {
		return nil, &models.Error{Code: http.StatusNotFound, Message: "Can't find forum"}
	}

	post := &models.Post{ID: info.ID}

	row = fr.dbConn.QueryRow(`UPDATE post SET message=COALESCE(NULLIF($1, ''), message),
                             isEdited = CASE WHEN $1 = '' OR message = $1 THEN isEdited ELSE true END
                             WHERE id=$2 RETURNING *`, info.Message, post.ID)
	err = row.Scan(&post.ID, &post.Author, &post.Created, &post.Forum,  &post.IsEdited,
		&post.Message, &post.Parent, &post.Thread, &post.Route)
	if err != nil {
		return post, &models.Error{Code: http.StatusInternalServerError, Message: err.Error()}
	}

	return post, nil
}

func (fr ForumRepository)PostInfo(id int, related models.Related) (*models.PostInfo, *models.Error){
	postAll := &models.PostInfo{}

	post := &models.Post{}
	query := `SELECT p.id, p.author, p.created, p.forum, p.isEdited, p.message, p.parent, p.thread`
	if related.IsForum {
		query += `, f.slug, f.author, f.title, f.posts, f.threads`
	}
	if related.IsUser{
		query +=`, u.nickname, u.fullname, u.about, u.email`
	}
	if related.IsThread{
		query +=`, t.id, t.title, t.author, t.forum, t.message, t.votes, t.slug, t.created`
	}

	query += ` FROM post AS p`
	if related.IsForum {
		query +=` JOIN forum AS f ON f.slug=p.forum`
	}
	if related.IsUser{
		query +=` JOIN users AS u ON u.nickname=p.author`
	}
	if related.IsThread{
		query +=` JOIN thread AS t ON t.id=p.thread`
	}
	query += ` WHERE p.id=$1`
	row := fr.dbConn.QueryRow(query, id)

	var params []interface{}
	params = append(params, &post.ID, &post.Author, &post.Created, &post.Forum,  &post.IsEdited,
		&post.Message, &post.Parent, &post.Thread)
	thread := &models.Thread{}
	if related.IsForum {
		forum := &models.Forum{}
		params = append(params, &forum.Slug, &forum.User, &forum.Title, &forum.Posts, &forum.Threads)
		postAll.Forum = forum
	}
	if related.IsUser{
		user := &models.User{}
		params = append(params, &user.Nickname, &user.FullName, &user.About, &user.Email)
		postAll.Author = user
	}
	if related.IsThread{
		params = append(params, &thread.ID, &thread.Title, &thread.Author, &thread.Forum, &thread.Message, &thread.Votes,
			&thread.Slug, &thread.Created)
	}

	postAll.Post = post
	err := row.Scan(params...)
	if(thread.Title != "") {
		if (thread.Slug == thread.Title+thread.Author) {
			postAll.Thread = models.DeleteSlug(thread)
		} else {
			postAll.Thread = thread
		}
	}
	if err != nil {
		return nil, &models.Error{Code: http.StatusNotFound, Message: "Can't find forum"}
	}

	return postAll, nil
}

func (fr ForumRepository)StatusDB() *models.Status{
	status := &models.Status{}
	err := fr.dbConn.QueryRow(`SELECT COUNT(*) FROM users;`).Scan(&status.User)
	if err != nil {
		status.User = 0
	}
	err = fr.dbConn.QueryRow(`SELECT COUNT(*) FROM forum;`).Scan(&status.Forum)
	if err != nil {
		status.Forum = 0
	}
	err = fr.dbConn.QueryRow(`SELECT COUNT(*) FROM thread;`).Scan(&status.Thread)
	if err != nil {
		status.Thread = 0
	}
	err = fr.dbConn.QueryRow(`SELECT COUNT(*) FROM post;`).Scan(&status.Post)
	if err != nil {
		status.Post = 0
	}
	return status
}

func (fr *ForumRepository) ClearDB() *models.Error {
	_, err := fr.dbConn.Exec(`TRUNCATE users, forum, thread, post, votes CASCADE;`)
	if err != nil {
		return nil
	}
	return nil
}
func (fr *ForumRepository)CreateUser(user *models.User)([]*models.User, *models.Error){
	var users []*models.User

	_, err := fr.dbConn.Exec(	`INSERT INTO users(nickname, fullname, about, email) VALUES ($1, $2, $3, $4);`,
		user.Nickname, user.FullName, user.About, user.Email)
	if err != nil {
		if err.(pgx.PgError).Code == "23505"{
			rows, err := fr.dbConn.Query(`SELECT nickname, fullName, about, email FROM users WHERE nickname=$1 or email=$2;`,
				user.Nickname, user.Email)
			defer rows.Close()
			if err != nil {
				return nil, &models.Error{Code: http.StatusInternalServerError, Message: err.Error()}
			}
			for rows.Next() {
				user := &models.User{}
				err := rows.Scan(&user.Nickname, &user.FullName, &user.About, &user.Email)
				if err != nil {
					return users, &models.Error{Code: http.StatusInternalServerError, Message: err.Error()}
				}
				users = append(users, user)
			}
			return users, &models.Error{Code: http.StatusConflict}
		} else {
			return nil, &models.Error{Code: http.StatusInternalServerError, Message: err.Error()}
		}
	}

	users = append(users, user)
	return users, nil
}

func (fr ForumRepository)GetUserProfile(nickname string) (*models.User, *models.Error){
	user := &models.User{}
	row := fr.dbConn.QueryRow(`SELECT nickname, fullname, about, email FROM users WHERE nickname=$1;`, nickname)
	err := row.Scan(&user.Nickname, &user.FullName, &user.About, &user.Email)
	if err != nil {
		return nil,  &models.Error{Code: http.StatusNotFound, Message: "Can't find forum"}
	}
	return user, nil
}

func (fr ForumRepository)UpdateUserProfile(user *models.User) *models.Error{
	err := fr.dbConn.QueryRow(
		`UPDATE users SET 
				email=COALESCE(NULLIF($1, ''), email), 
				about=COALESCE(NULLIF($2, ''), about),
				fullname=COALESCE(NULLIF($3, ''), fullname) 
				WHERE nickname=$4 RETURNING *`,
		user.Email, user.About, user.FullName, user.Nickname).Scan(&user.Nickname, &user.FullName, &user.About, &user.Email)
	if err != nil {
		if pgErr, ok := err.(pgx.PgError); ok && pgErr.Code == "23505" {
			return &models.Error{Code: http.StatusConflict, Message: "Can't find forum"}
		}
		return &models.Error{Code: http.StatusNotFound, Message: "Can't find forum"}
	}

	return nil
}

func (fr ForumRepository)CreatePosts(posts []*models.Post, slugOrID string) ([]*models.Post, *models.Error) {
	var threadID int
	var threadForum string
	var param interface{}
	query := `SELECT id, forum FROM thread WHERE `
	id, err := strconv.Atoi(slugOrID)
	if err != nil {
		query += `slug=$1`
		param = slugOrID
	} else {
		query += `id=$1`
		param = id
	}
	row := fr.dbConn.QueryRow(query, param)
	err = row.Scan(&threadID,  &threadForum)
	if err != nil {
		return nil, &models.Error{Code: http.StatusNotFound, Message: "Can't find forum"}
	}

	insertQuery := `INSERT INTO post(author, created, forum, message, parent, thread) VALUES `
	createTime := time.Now()
	var params []interface{}
	//var parent

	for i, post := range posts {
		if post.Parent == 0 {

		}
		insertQuery += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d),", i * 6 + 1, i * 6 + 2, i * 6 + 3, i * 6 + 4, i * 6 + 5, i * 6 + 6)
		params = append(params, post.Author, createTime, threadForum, post.Message, post.Parent, threadID)
	}

	insertQuery = strings.TrimSuffix(insertQuery, ",")
	insertQuery += ` RETURNING id, forum, isEdited, thread, created;`

	rows, err := fr.dbConn.Query(insertQuery, params...)
	if err != nil {
		if err.(pgx.PgError).Code == "23503" {
			return nil, &models.Error{Code: http.StatusNotFound, Message: "Can't find forum"}
		}
		return nil, &models.Error{Code: http.StatusConflict, Message: "Parent"}
	}

	defer rows.Close()

	for i, _ := range posts {
		if rows.Next() {
			err := rows.Scan(&(posts[i]).ID,&(posts[i]).Forum, &(posts[i]).IsEdited, &(posts[i]).Thread, &(posts[i]).Created)
			if err != nil {
				return nil, &models.Error{Code: http.StatusConflict, Message: "Can't find forum"}
			}
		}
	}

	if rows.Err() != nil {
		if rows.Err().(pgx.PgError).Code == "23503" {
			return nil, &models.Error{Code: http.StatusNotFound, Message: "Can't find forum"}
		}
		return nil, &models.Error{Code: http.StatusConflict, Message: "Parent"}
	}

	return posts, nil
}

func (fr ForumRepository)GetThreadInfo(slugOrID string) (*models.Thread, *models.Error){
	thread := &models.Thread{}
	var param interface{}
	query := `SELECT id, title, author, forum, message, votes, slug, created FROM thread WHERE `
	id, err := strconv.Atoi(slugOrID)
	if err != nil {
		query += `slug=$1`
		param = slugOrID
	} else {
		query += `id=$1`
		param = id
	}
	row := fr.dbConn.QueryRow(query, param)
	err = row.Scan(
		&thread.ID,
		&thread.Title,
		&thread.Author,
		&thread.Forum,
		&thread.Message,
		&thread.Votes,
		&thread.Slug,
		&thread.Created)
	if err != nil {
		return nil, &models.Error{Code: http.StatusNotFound, Message: "Can't find forum"}
	}

	return thread,nil
}

func (fr ForumRepository)UpdateThreadInfo(thread *models.Thread) *models.Error{
	query := `UPDATE thread SET 
				title=COALESCE(NULLIF($1, ''), title), 
				message=COALESCE(NULLIF($2, ''), message) 
				WHERE `
	params := []interface{}{thread.Title, thread.Message}

	if thread.Slug != "" {
		query += `slug=$3 `
		params = append(params, thread.Slug)
	} else {
		query += `id=$3 `
		params = append(params, thread.ID)
	}

	query += `RETURNING *`
	err := fr.dbConn.QueryRow(query, params...).Scan(&thread.ID, &thread.Title, &thread.Author, &thread.Created, &thread.Forum, &thread.Message, &thread.Slug, &thread.Votes)
	if err != nil {
		return &models.Error{Code: http.StatusNotFound, Message: "Can't find forum"}
	}

	return nil
}

func (fr *ForumRepository)	InsertOrUpdateVote(slugOrID string, vote *models.Vote) (*models.Thread, *models.Error){
	thread := &models.Thread{}
	var param interface{}
	query := `SELECT id FROM thread WHERE `
	id, err := strconv.Atoi(slugOrID)
	if err != nil {
		query += `slug=$1`
		param = slugOrID
	} else {
		query += `id=$1`
		param = id
	}
	row := fr.dbConn.QueryRow(query, param)
	err = row.Scan(
		&thread.ID)
	if err != nil {
		return nil, &models.Error{Code: http.StatusNotFound, Message: "Can't find forum"}
	}

	_, err = fr.dbConn.Exec(`INSERT INTO votes(author, voice, thread_id) VALUES ($1, $2, $3) ON CONFLICT (author, thread_id) DO UPDATE SET voice = $2;`, vote.Nickname,
		vote.Voice, thread.ID)
	if err != nil {
		return nil, &models.Error{Code: http.StatusNotFound, Message: "no user"}
	}

	err = fr.dbConn.QueryRow(`SELECT id, title, author, forum, message, votes, slug, created FROM thread WHERE id=$1`, thread.ID).Scan(
		&thread.ID,
		&thread.Title,
		&thread.Author,
		&thread.Forum,
		&thread.Message,
		&thread.Votes,
		&thread.Slug,
		&thread.Created)
	if err != nil {
		return nil, &models.Error{Code: http.StatusNotFound, Message: "Can't find forum"}
	}

	return thread, nil
}

func (fr ForumRepository)GetThreadPosts(slugOrID string, params *models.Params) ([]*models.Post, *models.Error){
	var threadID int
	var param interface{}
	query := `SELECT id FROM thread WHERE `
	id, err := strconv.Atoi(slugOrID)
	if err != nil {
		query += `slug=$1`
		param = slugOrID
	} else {
		query += `id=$1`
		param = id
	}
	err = fr.dbConn.QueryRow(query, param).Scan(&threadID)
	if err != nil {
		return nil, &models.Error{Code: http.StatusNotFound, Message: "Can't find forum"}
	}

	var selectPar []interface{}
	var posts []*models.Post

	query = `SELECT id, author, created, forum, isEdited, message, parent, thread FROM post`

	switch params.Sort {
	case "tree":
		if params.Since == "" {
			if params.Desc {
				query += ` WHERE thread=$1 ORDER BY route DESC, id DESC LIMIT $2;`
			} else {
				query += ` WHERE thread=$1 ORDER BY route ASC, id  ASC LIMIT $2;`
			}
			selectPar = append(selectPar, threadID, params.Limit)
		} else {
			if params.Desc {
				query += ` WHERE thread=$1 AND ROUTE < (SELECT route FROM post WHERE id = $2)
		ORDER BY route DESC, id  DESC LIMIT $3;`

			} else {
				query += ` WHERE thread=$1 AND ROUTE > (SELECT route FROM post WHERE id = $2)
		ORDER BY route ASC, id  ASC LIMIT $3;`
			}
			selectPar = append(selectPar, threadID, params.Since, params.Limit)
		}
	case "parent_tree":
		if params.Since == "" {
			if params.Desc {
				query += ` WHERE route[1] IN (SELECT id FROM post WHERE thread = $1 AND parent = 0 ORDER BY id DESC LIMIT $2)
			ORDER BY route[1] DESC, route, id;`
			} else {
				query +=` WHERE route[1] IN (SELECT id FROM post WHERE thread = $1 AND parent = 0 ORDER BY id LIMIT $2)
			ORDER BY route, id;`
			}
			selectPar =append(selectPar, threadID, params.Limit)
		} else {
			if params.Desc {
				query += ` WHERE route[1] IN (SELECT id FROM post WHERE thread = $1 AND parent = 0 AND ROUTE[1] <
				(SELECT route[1] FROM post WHERE id = $2) ORDER BY id DESC LIMIT $3) ORDER BY route[1] DESC, route, id;`
			} else {
				query += ` WHERE route[1] IN (SELECT id FROM post WHERE thread = $1 AND parent = 0 AND ROUTE[1] >
				(SELECT route[1] FROM post WHERE id = $2) ORDER BY id ASC LIMIT $3) ORDER BY route, id;`
			}
			selectPar =append(selectPar, threadID, params.Since, params.Limit)
		}
	default:
		if params.Since == "" {
			if params.Desc {
				query +=` WHERE thread=$1 ORDER BY id DESC LIMIT $2;`
			} else {
				query +=` WHERE thread=$1 ORDER BY id LIMIT $2;`
			}
			selectPar =append(selectPar, threadID, params.Limit)
		} else {
			if params.Desc {
				query += ` WHERE thread=$1 AND id < $2 ORDER BY id DESC LIMIT $3;`
			} else {
				query +=` WHERE thread=$1 AND id > $2 ORDER BY id LIMIT $3;`
			}
			selectPar =append(selectPar, threadID, params.Since, params.Limit)
		}
	}
	rows, err := fr.dbConn.Query(query, selectPar...)
	if err != nil {
		return nil, &models.Error{Code: http.StatusInternalServerError, Message: err.Error()}
	}
	defer rows.Close()

	for rows.Next() {
		post := &models.Post{}
		err = rows.Scan(&post.ID, &post.Author, &post.Created, &post.Forum, &post.IsEdited, &post.Message, &post.Parent, &post.Thread)
		if err != nil {
			return nil, &models.Error{Code: http.StatusInternalServerError, Message: err.Error()}
		}

		posts = append(posts, post)
	}
	return posts, nil
}