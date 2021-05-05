package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
)

func main() {
	// Create a Note object
	var note vocab.ActivityStreamsNote = streams.NewActivityStreamsNote()

	// Create an `id` property and set it on the Note
	id, _ := url.Parse("https://example.com/some/path/to/this/note")
	var idProperty vocab.JSONLDIdProperty = streams.NewJSONLDIdProperty()
	idProperty.Set(id)

	// Set the `id` property on our Note.
	note.SetJSONLDId(idProperty)

	// Let's try to add content to our note. First, let's get the property.
	contentProperty := note.GetActivityStreamsContent()

	// WARNING: Missing properties are `nil`!
	if contentProperty == nil {
		// Create a new property and set it on the note.
		contentProperty = streams.NewActivityStreamsContentProperty()
		// Treat properties as pointers, not values. Setting a
		// property is not a value-copy so if we modify
		// the property later, any modification will be
		// reflected in the note.
		note.SetActivityStreamsContent(contentProperty)
	}

	// Now we are guaranteed a non-`nil` property: let's add content!
	contentProperty.AppendXMLSchemaString("Hello, world!")

	// The "published" property is functional: It can only have at most one value.
	published := streams.NewActivityStreamsPublishedProperty()
	// We can set a time...
	published.Set(time.Now())
	// ...or, in this very unusual practice, set it as an IRI
	// iri, _ := url.Parse("https://go-fed.org/some/path")
	// published.SetIRI(iri)

	if published.IsIRI() {
		fmt.Println(published.GetIRI())
	} else if published.IsXMLSchemaDateTime() {
		fmt.Println(published.Get())
	}

	// The "object" property is non-functional: It can have many values.
	object := streams.NewActivityStreamsObjectProperty()
	// We can append...
	object.AppendActivityStreamsNote(note)
	// ...or prepend...
	object.PrependActivityStreamsArticle(streams.NewActivityStreamsArticle())
	// ...and IRIs too
	iri, _ := url.Parse("https://go-fed.org/foo")
	object.AppendIRI(iri)

	// An iterator interface lets you work with each element
	// for iter := object.Begin(); iter != iter.End(); iter = iter.Next() {
	// 	fmt.Println(iter.KindIndex())
	// 	if iter.IsActivityStreamsNote() {
	// 		note := iter.GetActivityStreamsNote()
	// 		fmt.Println(note)
	// 	} else if published.IsXMLSchemaDateTime() {
	// 		article := iter.GetActivityStreamsArticle()
	// 		fmt.Println(article)
	// 	} else if published.IsIRI() {
	// 		iri := iter.GetIRI()
	// 		fmt.Println(iri)
	// 	}
	// }

	// Deserialize a JSON payload
	jsonstr := `{
		"@context": "https://www.w3.org/ns/activitystreams",
		"id":       "https://go-fed.org/foo",
		"name":     "Foo Bar",
		"inbox":    "https://go-fed.org/foo/inbox",
		"outbox":   "https://go-fed.org/foo/outbox",
		"type":     "Person",
		"url":      "https://go-fed.org/foo"
	}`
	var m map[string]interface{}
	_ = json.Unmarshal([]byte(jsonstr), &m)

	// Next, we prepare a streams.JSONResolver, providing one or more callbacks.
	var person vocab.ActivityStreamsPerson
	resolver, _ := streams.NewJSONResolver(func(c context.Context, p vocab.ActivityStreamsPerson) error {
		// Store the person in the enclosing scope, for later.
		person = p
		return nil
	}, func(c context.Context, note vocab.ActivityStreamsNote) error {
		// We can treat the type differently.
		fmt.Println(note)
		return nil
	})
	// It will call back a function we provide if it is of a matching type,
	// or returns streams.ErrNoCallbackMatch when we didn't give it a matcher for
	// the type, or streams.ErrUnhandledType if it is a type unknown to Go-Fed.
	ctx := context.Background()
	err := resolver.Resolve(ctx, m)
	if err != nil {
		fmt.Println("Error: " + err.Error())
	}
	// fmt.Println(person)

	// Serialize to a JSON payload
	var jsonmap map[string]interface{}
	jsonmap, _ = streams.Serialize(person) // WARNING: Do not call the Serialize() method on person
	fmt.Println(jsonmap)
	_, _ = json.Marshal(jsonmap)

	// vocab.Type represents an ActivityStreams type in general
	var aType vocab.ActivityStreamsCollection = streams.NewActivityStreamsCollection()
	// ...which can be used for serializing...
	jsonmap, _ = streams.Serialize(aType)
	_, _ = json.Marshal(jsonmap)

	// ...or branching logic based on its precise type...
	typeResolver, _ := streams.NewTypeResolver(func(c context.Context, oc vocab.ActivityStreamsOrderedCollection) error {
		fmt.Println(oc)
		return nil
	}, func(c context.Context, oc vocab.ActivityStreamsCollection) error {
		fmt.Println(c)
		return nil
	})
	// This is a TypeResolver, not a JSONResolver, so it accepts a vocab.Type
	// instead of a map[string]interface{}.
	ctx = context.Background()
	_ = typeResolver.Resolve(ctx, aType)
	// fmt.Println(aType.GetActivityStreamsFirst().Name())

}
