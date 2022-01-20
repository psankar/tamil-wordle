package dao

import (
	"context"
	"fmt"

	firestore "cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/iterator"
)

type User struct {
	Id            string
	Name          string
	TwitterHandle string
	PublicKey     string
	Active        bool
}

type Word struct {
	Id     string
	Word   string
	Date   string
	UserId string
}

const usersCollectionName = "users"
const wordsCollectionName = "words"

func openClient() (context.Context, *firestore.Client, error) {
	ctx := context.Background()
	app, err := firebase.NewApp(ctx, nil)
	if err != nil {
		err = fmt.Errorf("error initializing app: %v", err)
		return nil, nil, err
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		err = fmt.Errorf("error initializing Firestore: %v", err)
		return nil, nil, err
	}
	return ctx, client, nil
}

func ListUsers() ([]User, error) {
	ctx, client, err := openClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()
	iter := client.Collection(usersCollectionName).Documents(ctx)
	results := []User{}
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			err = fmt.Errorf("failed to iterate: %v", err)
			return nil, err
		}
		var userObj User
		doc.DataTo(&userObj)
		userObj.Id = doc.Ref.ID
		results = append(results, userObj)
	}
	return results, nil
}

func CreateUser(user User) (string, error) {
	ctx, client, err := openClient()
	if err != nil {
		return "", err
	}
	defer client.Close()

	ref := client.Collection(usersCollectionName).NewDoc()
	user.Id = ref.ID
	_, err = ref.Set(ctx, user)
	if err != nil {
		err = fmt.Errorf("failed to add user: %v", err)
		return "", err
	}
	return ref.ID, err
}

func MarkUserActive(id string) error {
	ctx, client, err := openClient()
	if err != nil {
		return err
	}
	defer client.Close()

	// Inactivate existing user
	iter := client.Collection(usersCollectionName).Where("Active", "==", true).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to iterate existing active users: %v", err)
		}
		var userObj User
		doc.DataTo(&userObj)
		userObj.Id = doc.Ref.ID
		userObj.Active = false
		_, err = doc.Ref.Set(ctx, userObj)
		if err != nil {
			return fmt.Errorf("failed to inactivate existing active user: %v", err)
		}
	}

	// Activate current user
	doc, err := client.Collection(usersCollectionName).Doc(id).Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get user: %v", err)
	}
	var userObj User
	doc.DataTo(&userObj)
	userObj.Id = doc.Ref.ID
	userObj.Active = true
	_, err = doc.Ref.Set(ctx, userObj)
	if err != nil {
		return fmt.Errorf("failed to activate user: %v", err)
	}
	return nil
}

func UpdatePublicKey(id string, publicKey string) error {
	ctx, client, err := openClient()
	if err != nil {
		return err
	}
	defer client.Close()

	doc, err := client.Collection(usersCollectionName).Doc(id).Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get user: %v", err)
	}
	var userObj User
	doc.DataTo(&userObj)
	userObj.PublicKey = publicKey
	_, err = doc.Ref.Set(ctx, userObj)
	if err != nil {
		return fmt.Errorf("failed to update public key: %v", err)
	}
	return nil
}

func AddWord(word Word) (string, error) {
	ctx, client, err := openClient()
	if err != nil {
		return "", err
	}
	defer client.Close()

	// Check if word already exists for the day
	w, err := GetWordForTheDay(word.Date)
	if err != nil {
		err = fmt.Errorf("failed to check if word exists for the day: %v", err)
		return "", err
	}
	if w.Id != "" {
		err = fmt.Errorf("word already exists for the day")
		return "", err
	}

	ref := client.Collection(wordsCollectionName).NewDoc()
	word.Id = ref.ID
	_, err = ref.Set(ctx, word)
	if err != nil {
		err = fmt.Errorf("failed to add word: %v", err)
		return "", err
	}
	return ref.ID, err
}

func GetWordForTheDay(date string) (Word, error) {
	ctx, client, err := openClient()
	if err != nil {
		return Word{}, err
	}
	defer client.Close()

	iter := client.Collection(wordsCollectionName).Where("Date", "==", date).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return Word{}, fmt.Errorf("failed to iterate: %v", err)
		}
		var wordObj Word
		doc.DataTo(&wordObj)
		wordObj.Id = doc.Ref.ID
		if wordObj.Word != "" {
			return wordObj, nil
		}
	}
	return Word{}, nil
}