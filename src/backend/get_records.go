package main

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func getRecords(collection *mongo.Collection, ctx context.Context) (map[string]interface{}, error) {

	cur, err := collection.Find(ctx, bson.D{})

	if err != nil {
		return nil, err
	}

	defer cur.Close(ctx)

	var users []bson.M

	for cur.Next(ctx) {

		var user bson.M

		if err = cur.Decode(&user); err != nil {
			return nil, err
		}

		users = append(users, user)

	}

	res := map[string]interface{}{}

	res = map[string]interface{}{
		"data": users,
	}

	return res, nil
}
