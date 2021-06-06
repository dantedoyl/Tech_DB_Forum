package delivery

import (
	"encoding/json"
	"fmt"
	"github.com/dantedoyl/Tech_DB_Forum/internal/forum"
	"github.com/dantedoyl/Tech_DB_Forum/internal/models"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"net/http"
	"strconv"
	"strings"
)

type ForumHandler struct {
	ForumRepo forum.ForumRepository
}

func NewForumHandler(r *mux.Router, forumRepo forum.ForumRepository)  *ForumHandler{
	fh := &ForumHandler{ForumRepo: forumRepo}
	r.HandleFunc("/forum/create", fh.CreateForum).Methods(http.MethodPost)
	r.HandleFunc("/forum/{slug}/details", fh.ForumInfo).Methods(http.MethodGet)
	r.HandleFunc("/forum/{slug}/create", fh.CreateThread).Methods(http.MethodPost)
	r.HandleFunc("/forum/{slug}/users", fh.ForumUsers).Methods(http.MethodGet)
	r.HandleFunc("/forum/{slug}/threads", fh.ForumThreads).Methods(http.MethodGet)
	r.HandleFunc("/post/{id}/details", fh.PostInfoUpdate).Methods(http.MethodPost)
	r.HandleFunc("/post/{id}/details", fh.PostInfo).Methods(http.MethodGet)
	r.HandleFunc("/service/status", fh.StatusDB).Methods(http.MethodGet)
	r.HandleFunc("/service/clear", fh.ClearDB).Methods(http.MethodPost)
	r.HandleFunc("/thread/{slug_or_id}/create", fh.CreatePost).Methods(http.MethodPost)
	r.HandleFunc("/thread/{slug_or_id}/details", fh.ThreadInfo).Methods(http.MethodGet)
	r.HandleFunc("/thread/{slug_or_id}/details", fh.UpdateThread).Methods(http.MethodPost)
	r.HandleFunc("/thread/{slug_or_id}/posts", fh.ThreadPosts).Methods(http.MethodGet)
	r.HandleFunc("/thread/{slug_or_id}/vote", fh.Vote).Methods(http.MethodPost)
	r.HandleFunc("/user/{nickname}/create", fh.CreateUser).Methods(http.MethodPost)
	r.HandleFunc("/user/{nickname}/profile", fh.UserProfile).Methods(http.MethodGet)
	r.HandleFunc("/user/{nickname}/profile", fh.UserProfileUpdate).Methods(http.MethodPost)
	return fh
}

func (fh *ForumHandler) CreateForum(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	forum := &models.Forum{}
	err := json.NewDecoder(r.Body).Decode(forum)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	er := fh.ForumRepo.CreateForum(forum)
	if er != nil{
		if er.Code == http.StatusConflict{
			w.WriteHeader(http.StatusConflict)
			body, err := json.Marshal(forum)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write(models.ErrorToJSON(er.Message))
				return
			}
			w.Write(body)
			return
		}
		w.WriteHeader(er.Code)
		w.Write(models.ErrorToJSON(er.Message))
		return
	}

	w.WriteHeader(http.StatusCreated)
	body, err := json.Marshal(forum)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(models.ErrorToJSON(err.Error()))
		return
	}
	w.Write(body)
}

func (fh *ForumHandler) ForumInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	slug := vars["slug"]

	forum, er := fh.ForumRepo.GetForumInfo(slug)
	if er != nil {
		w.WriteHeader(er.Code)
		w.Write(models.ErrorToJSON(er.Message))
		return
	}
	w.WriteHeader(http.StatusOK)
	body, err := json.Marshal(forum)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(models.ErrorToJSON(err.Error()))
		return
	}
	w.Write(body)
}

func (fh *ForumHandler) CreateThread(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	slug := vars["slug"]

	thread := &models.Thread{}
	err := json.NewDecoder(r.Body).Decode(thread)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	thread.Forum = slug
	er := fh.ForumRepo.CreateThread(thread)
	if er != nil{
		if er.Code == http.StatusConflict{
			w.WriteHeader(http.StatusConflict)
			body, err := json.Marshal(thread)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write(models.ErrorToJSON(er.Message))
				return
			}
			w.Write(body)
			return
		}
		w.WriteHeader(er.Code)
		w.Write(models.ErrorToJSON(er.Message))
		return
	}
	w.WriteHeader(http.StatusCreated)
	var body []byte
	if(thread.Slug == ""){
		body, err = json.Marshal(models.DeleteSlug(thread))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(models.ErrorToJSON(err.Error()))
			return
		}
	} else {
		body, err = json.Marshal(thread)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(models.ErrorToJSON(err.Error()))
			return
		}
	}
	w.Write(body)
}

func (fh *ForumHandler)ForumUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := &models.Params{}
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	_ = decoder.Decode(params, r.URL.Query())
	fmt.Println(params)
	vars := mux.Vars(r)
	slug := vars["slug"]

	users, er := fh.ForumRepo.GetForumUsers(slug, params)
	if er != nil {
		w.WriteHeader(er.Code)
		w.Write(models.ErrorToJSON(er.Message))
		return
	}

	w.WriteHeader(http.StatusOK)
	if len(users) == 0 {
		w.Write([]byte("[]"))
		return
	}
	body, err := json.Marshal(users)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(models.ErrorToJSON(err.Error()))
		return
	}
	w.Write(body)
}

func (fh *ForumHandler)ForumThreads(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params :=  &models.Params{}
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	_ = decoder.Decode(params, r.URL.Query())

	vars := mux.Vars(r)
	slug := vars["slug"]


	threads, er := fh.ForumRepo.GetForumThreads(slug, params)
	if er != nil {
		w.WriteHeader(er.Code)
		w.Write(models.ErrorToJSON(er.Message))
		return
	}

	if len(threads) == 0 {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("[]"))
		return
	}

	var thr []interface{}
	for _, thread := range threads{
		if (thread.Slug == thread.Title + thread.Author){
			thr = append(thr, models.DeleteSlug(thread))
		} else {
			thr = append(thr, thread)
		}
	}

	w.WriteHeader(http.StatusOK)

	body, err := json.Marshal(thr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(models.ErrorToJSON(err.Error()))
		return
	}
		w.Write(body)
}

func (fh *ForumHandler) PostInfoUpdate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	postUpdate := &models.PostUpdate{}
	postUpdate.ID = id
	err := json.NewDecoder(r.Body).Decode(postUpdate)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	post, er := fh.ForumRepo.UpdatePostInfo(postUpdate)
	if er != nil {
		w.WriteHeader(er.Code)
		w.Write(models.ErrorToJSON(er.Message))
		return
	}

	body, err := json.Marshal(post)
	if err != nil {

		w.WriteHeader(http.StatusInternalServerError)
		w.Write(models.ErrorToJSON(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func (fh *ForumHandler) PostInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	relatedString := r.URL.Query().Get("related")
	related := models.Related{}

	if strings.Contains(relatedString, "user") {
		related.IsUser = true
	}

	if strings.Contains(relatedString, "thread") {
		related.IsThread = true
	}

	if strings.Contains(relatedString, "forum") {
		related.IsForum = true
	}

	post, er := fh.ForumRepo.PostInfo(id, related)
	if er != nil {
		w.WriteHeader(er.Code)
		w.Write(models.ErrorToJSON(er.Message))
		return
	}

	body, err := json.Marshal(post)
	if err != nil {

		w.WriteHeader(http.StatusInternalServerError)
		w.Write(models.ErrorToJSON(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func (fh *ForumHandler) StatusDB(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	status := fh.ForumRepo.StatusDB()
	body, err := json.Marshal(status)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(models.ErrorToJSON(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func (fh *ForumHandler) ClearDB(w http.ResponseWriter, r *http.Request)  {
	w.Header().Set("Content-Type", "application/json")

	er := fh.ForumRepo.ClearDB()
	if er != nil {
		w.WriteHeader(er.Code)
		w.Write(models.ErrorToJSON(er.Message))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (fh *ForumHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	nickname, _ := vars["nickname"]

	user := &models.User{}
	err := json.NewDecoder(r.Body).Decode(user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user.Nickname = nickname

	users, er := fh.ForumRepo.CreateUser(user)
	if er != nil{
		if er.Code == http.StatusConflict {
			w.WriteHeader(er.Code)
			body, err := json.Marshal(users)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write(body)
				return
			}
			w.Write(body)
			return
		}
			w.WriteHeader(er.Code)
			w.Write(models.ErrorToJSON(er.Message))
			return
	}

	w.WriteHeader(http.StatusCreated)
	body, err := json.Marshal(users[0])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(models.ErrorToJSON(er.Message))
		return
	}
	w.Write(body)
}

func (fh *ForumHandler) UserProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	nickname, _ := vars["nickname"]

	user, er := fh.ForumRepo.GetUserProfile(nickname)
	if er != nil {
		w.WriteHeader(er.Code)
		w.Write(models.ErrorToJSON(er.Message))
		return
	}

	body, err := json.Marshal(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(models.ErrorToJSON(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func (fh *ForumHandler) UserProfileUpdate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	nickname, _ := vars["nickname"]

	user := &models.User{}
	err := json.NewDecoder(r.Body).Decode(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(models.ErrorToJSON(err.Error()))
		return
	}

	user.Nickname = nickname

	er := fh.ForumRepo.UpdateUserProfile(user)
	if er != nil {
		w.WriteHeader(er.Code)
		w.Write(models.ErrorToJSON(er.Message))
		return
	}

	body, err := json.Marshal(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(models.ErrorToJSON(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func (fh *ForumHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	slugOrID, _ := vars["slug_or_id"]

	var posts []*models.Post
	err := json.NewDecoder(r.Body).Decode(&posts)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, er := fh.ForumRepo.CreatePosts(posts, slugOrID)
	if er != nil {
		w.WriteHeader(er.Code)
		w.Write(models.ErrorToJSON(er.Message))
		return
	}

	if len(posts) == 0 {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("[]"))
		return
	}

	body, err := json.Marshal(posts)
	if err != nil {

		w.WriteHeader(http.StatusInternalServerError)
		w.Write(models.ErrorToJSON(err.Error()))
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(body)
}

func (fh *ForumHandler) ThreadInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	slugOrID, _ := vars["slug_or_id"]

	thread, er := fh.ForumRepo.GetThreadInfo(slugOrID)
	if er != nil {
		w.WriteHeader(er.Code)
		w.Write(models.ErrorToJSON(er.Message))
		return
	}

	var body []byte
	var err error
	if (thread.Slug == thread.Title+thread.Author) {
		body, err = json.Marshal(models.DeleteSlug(thread))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(models.ErrorToJSON(err.Error()))
			return
		}
	} else {
		body, err = json.Marshal(thread)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(models.ErrorToJSON(err.Error()))
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func (fh *ForumHandler) UpdateThread(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	slugOrID, _ := vars["slug_or_id"]

	thread := &models.Thread{}
	err := json.NewDecoder(r.Body).Decode(&thread)
	if err != nil {

		w.WriteHeader(http.StatusInternalServerError)
		w.Write(models.ErrorToJSON(err.Error()))
		return
	}

	id, err := strconv.Atoi(slugOrID)
	if err != nil {
		thread.Slug = slugOrID
	} else {
		thread.ID = id
	}

	er := fh.ForumRepo.UpdateThreadInfo(thread)
	if er != nil {
		w.WriteHeader(er.Code)
		w.Write(models.ErrorToJSON(er.Message))
		return
	}

	var body []byte
	if (thread.Slug == thread.Title+thread.Author) {
		body, err = json.Marshal(models.DeleteSlug(thread))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(models.ErrorToJSON(err.Error()))
			return
		}
	} else {
		body, err = json.Marshal(thread)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(models.ErrorToJSON(err.Error()))
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func (fh *ForumHandler) Vote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	slugOrID, _ := vars["slug_or_id"]

	vote := &models.Vote{}
	err := json.NewDecoder(r.Body).Decode(&vote)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(models.ErrorToJSON(err.Error()))
		return
	}

	thread, er := fh.ForumRepo.InsertOrUpdateVote(slugOrID, vote)
	if er != nil {
		w.WriteHeader(er.Code)
		w.Write(models.ErrorToJSON(er.Message))
		return
	}

	var body []byte
	if (thread.Slug == thread.Title+thread.Author) {
		body, err = json.Marshal(models.DeleteSlug(thread))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(models.ErrorToJSON(err.Error()))
			return
		}
	} else {
		body, err = json.Marshal(thread)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(models.ErrorToJSON(err.Error()))
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func (fh *ForumHandler) ThreadPosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	slugOrID, _ := vars["slug_or_id"]

	params :=  &models.Params{}
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	_ = decoder.Decode(params, r.URL.Query())

	posts, er := fh.ForumRepo.GetThreadPosts(slugOrID, params)
	if er != nil {
		w.WriteHeader(er.Code)
		w.Write(models.ErrorToJSON(er.Message))
		return
	}

	if len(posts) == 0 {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("[]"))
		return
	}

	body, err := json.Marshal(posts)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(models.ErrorToJSON(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(body)
}