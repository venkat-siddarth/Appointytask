package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	createusers = regexp.MustCompile(`^\/users[\/]*$`)
	getusers    = regexp.MustCompile(`^\/users[\/]+(\w+)[\/]*$`)
	createposts = regexp.MustCompile(`\/posts[\/]*$`)
	allposts    = regexp.MustCompile(`^\/posts\/users\/(\w+)$`)
	getposts    = regexp.MustCompile(`^\/posts\/(\w+)$`)
	client      *mongo.Client
)

type users struct {
	ID       primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name     string             `json:"name,omitempty" bson:"name,omitempty"`
	Email    string             `json:"email,omitempty" bson:"email,omitempty"`
	Password string             `json:"password,omitempty" bson:"password,omitempty"`
}
type posts struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	UserID    string             `json:"userid,omitempty" bson:"userid,omitempty"`
	Caption   string             `json:"caption,omitempty" bson:"caption,omitempty"`
	ImgURL    string             `json:"imgurl,omitempty" bson:"imgurl,omitempty"`
	Timestamp time.Time          `json:"tmpstamp,omitempty" bson:"tmpstamp,omitempty"`
}
type userHandler struct {
	*sync.RWMutex
}

func (h *userHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var userobj users
	w.Header().Add("content-type", "application/json")
	err := json.NewDecoder(r.Body).Decode(&userobj)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	collection := client.Database("UsersData").Collection("users")
	collection.InsertOne(ctx, userobj)
	if err != nil {
		notFound(w, r)
		return
	}

}
func (h *userHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	matches := getusers.FindStringSubmatch(r.URL.Path)
	fmt.Println(matches)
	if len(matches) < 2 {
		notFound(w, r)
		return
	}
	user_id := matches[1]
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	collection := client.Database("UsersData").Collection("users")
	objId, _ := primitive.ObjectIDFromHex(user_id)
	result := collection.FindOne(ctx, bson.M{"_id": objId})
	var parsedData bson.M
	if err := result.Decode(&parsedData); err != nil {

		notFound(w, r)
	}
	parsedString, _ := json.Marshal(parsedData)
	fmt.Fprint(w, string(parsedString))

}
func (h *userHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	//defer client.
	var postobj posts
	w.Header().Add("content-type", "application/json")
	err := json.NewDecoder(r.Body).Decode(&postobj)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	collection := client.Database("UsersData").Collection("posts")
	collection.InsertOne(ctx, postobj)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func (h *userHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	matches := getposts.FindStringSubmatch(r.URL.Path)
	fmt.Println(matches)
	if len(matches) < 2 {
		notFound(w, r)
		return
	}
	post_id := matches[1]
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	collection := client.Database("UsersData").Collection("posts")
	objId, _ := primitive.ObjectIDFromHex(post_id)
	result := collection.FindOne(ctx, bson.M{"_id": objId})
	var parsedData bson.M
	if err := result.Decode(&parsedData); err != nil {
		notFound(w, r)
	}
	parsedString, _ := json.Marshal(parsedData)
	fmt.Fprint(w, string(parsedString))

}

func (h *userHandler) AllPosts(w http.ResponseWriter, r *http.Request) {
	matches := allposts.FindStringSubmatch(r.URL.Path)
	fmt.Println(matches)
	if len(matches) < 2 {
		notFound(w, r)
		return
	}
	user_id := matches[1]
	fmt.Println(user_id)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	collection := client.Database("UsersData").Collection("posts")
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		notFound(w, r)
	}
	if err == nil {
		var parsedData []bson.M
		if err := cursor.All(ctx, &parsedData); err != nil {
			notFound(w, r)
		}
		parsedString, _ := json.Marshal(parsedData)
		fmt.Fprint(w, string(parsedString))

	}
}
func notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(`{"error":"not found"}`))
}
func (h *userHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	switch {
	case r.Method == http.MethodPost && createusers.MatchString(r.URL.Path):
		h.CreateUser(w, r)
		return
	case r.Method == http.MethodGet && getusers.MatchString(r.URL.Path):
		h.GetUser(w, r)
		return
	case r.Method == http.MethodPost && createposts.MatchString(r.URL.Path):
		h.CreatePost(w, r)
		return
	case r.Method == http.MethodGet && allposts.MatchString(r.URL.Path):
		h.AllPosts(w, r)
		return
	case r.Method == http.MethodGet && getposts.MatchString(r.URL.Path):
		h.GetPost(w, r)
		return
	default:
		notFound(w, r)
		return
	}
}

/*func homePage(w http.ResponseWriter, r *http.Request) {
	body :=
		fmt.Fprintf(w)

}*/

func main() {
	var testUser users
	testUser.Email = "default"
	testUser.Name = "default"
	testUser.Password = "default"
	mux := http.NewServeMux()
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, _ = mongo.Connect(ctx, clientOptions)
	server := &userHandler{}
	mux.Handle("/", server)
	fmt.Println("Mongo connection successful")
	mux.Handle("/users/", server)
	mux.Handle("/users", server)
	mux.Handle("/posts", server)
	mux.Handle("/posts/", server)
	mux.Handle("/posts/users/", server)
	http.ListenAndServe(":3000", server)

}
