package guestbook

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
  "log"
	"net/http"
	"strconv"

	"appengine"
	"appengine/datastore"
	"appengine/user"
  "appengine/urlfetch"

)

const (
	DogKeyKind = "Dog"
)

var (
	userWhitelist = map[string]bool{
		"test@example.com": true,
	}
)

type Dog struct {
	EntityId string `json:"entity_id"`
	Name     string `json:"name"`
	Gender   string `json:"gender"`
	Age      string `json:"age"`
	Breed    string `json:"breed"`
	// This dog's background, story, etc.
	Bio string `json:"bio"`
	// FacebookAlbum is the URL of a Facebook Album containing pictures of this dog.
	FacebookAlbumId string `json:"facebook_album_id"`
  Images []string `json:"images"`
	Adopted         bool   `json:"adopted"`
}

func init() {
	http.HandleFunc("/api/albums", albumsHandler)
	http.HandleFunc("/api/dogs", dogsHandler)
}

func albumsHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
  client := urlfetch.Client(c)
  id := r.FormValue("id")
  if len(id) == 0 {
    http.Error(w, "ID is required.", http.StatusBadRequest)
    return
  }
  access_token := "foo"
  url := "https://graph.facebook.com/v2.8/" + id + "?fields=photos{images}&access_token=" + access_token
  log.Printf("Attempting to fetch: %s", url)
  resp, err := client.Get(url)
  if err != nil {
    log.Print("Failed to fetch.")
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  if resp.StatusCode != http.StatusOK {
    log.Print("Error fetching " + url)
    http.Error(w, "Facebook returned status " + strconv.Itoa(resp.StatusCode), http.StatusBadRequest)
    return
  }
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  _, err = w.Write(body)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
}

func dogsHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	switch r.Method {
	case "POST":
		u := user.Current(c)
		// Enforce that the user is logged in.
		if u == nil {
			url, err := user.LoginURL(c, r.URL.String())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Location", url)
			w.WriteHeader(http.StatusFound)
			return
		}
		// Enforce that user is on whitelist.
		if !userWhitelist[u.String()] {
			http.Error(w, "User is not whitelisted", http.StatusForbidden)
			return
		}
		addDog(c, w, r)
	case "GET":
		fallthrough
	case "":
		listDogs(c, w, r)
	default:
		http.Error(w, "Not supported yet", http.StatusNotImplemented)
	}
}

func listDogs(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	var dogs []Dog
	// TODO: consider using an ancestor query for strong consistency.
	keys, err := datastore.NewQuery(DogKeyKind).GetAll(c, &dogs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
  if len(dogs) == 0 {
    fmt.Fprintf(w, "[]")
    return
  }
	if len(keys) != len(dogs) {
		http.Error(w, "Keys and dogs don't line up", http.StatusInternalServerError)
	}
	for i := 0; i < len(dogs); i++ {
		dogs[i].EntityId = keys[i].Encode()
	}
	resp, err := json.Marshal(dogs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	fmt.Fprintf(w, "%s\n", resp)
}

func addDog(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	if len(name) == 0 {
		http.Error(w, "Name is required.", http.StatusBadRequest)
		return
	}
	gender := r.FormValue("gender")
	if len(gender) == 0 {
		http.Error(w, "Gender is required.", http.StatusBadRequest)
		return
	}
	age := r.FormValue("age")
	if len(age) == 0 {
		http.Error(w, "Age is required.", http.StatusBadRequest)
		return
	}
	breed := r.FormValue("breed")
	if len(breed) == 0 {
		http.Error(w, "Breed is required.", http.StatusBadRequest)
		return
	}
	bio := r.FormValue("bio")
	if len(bio) == 0 {
		http.Error(w, "Bio is required.", http.StatusBadRequest)
		return
	}
	facebook_album_id := r.FormValue("facebook_album_id")
	if len(facebook_album_id) == 0 {
		http.Error(w, "Album ID is required.", http.StatusBadRequest)
		return
	}
  var images []string
  images_str := r.FormValue("images")
  if len(images_str) == 0 {
    images = make([]string, 0);
  } else {
    r.ParseForm()
    images = r.Form["images"]
  }
	adoptedstr := r.FormValue("adopted")
	if len(adoptedstr) == 0 {
		adoptedstr = "false"
	}
	if adoptedstr == "on" {
		adoptedstr = "true"
	}
	adopted, err := strconv.ParseBool(adoptedstr)
	if err != nil {
		http.Error(w, "Adopted is malformed.", http.StatusBadRequest)
		return
	}
	d := Dog{
		Name:            name,
		Gender:          gender,
		Age:             age,
		Breed:           breed,
		Bio:             bio,
		FacebookAlbumId: facebook_album_id,
    Images: images,
		Adopted:         adopted,
	}

	var key *datastore.Key
	entity_id := r.FormValue("entity_id")
	if len(entity_id) == 0 {
		key = datastore.NewIncompleteKey(c, DogKeyKind, nil)
	} else {
		key, err = datastore.DecodeKey(entity_id)
		if err != nil {
			http.Error(w, "Failed to parse key.", http.StatusBadRequest)
			return
		}
	}

	_, err = datastore.Put(c, key, &d)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}
