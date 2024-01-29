package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Database struct {
	*mongo.Database
}

type Collection struct {
	*mongo.Collection
}

func NewDatabase(ctx context.Context, uri string, name string) (*Database, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return &Database{
		Database: client.Database(name),
	}, nil
}

func (db *Database) Close(ctx context.Context) error {
	return db.Database.Client().Disconnect(ctx)
}

func (db *Database) Collection(name string) *Collection {
	return &Collection{
		Collection: db.Database.Collection(name),
	}
}

func (c *Collection) Insert(ctx context.Context, document interface{}) error {
	if _, err := c.Collection.InsertOne(ctx, document); err != nil {
		return err
	}

	return nil
}

func (c *Collection) InsertOne(ctx context.Context, document interface{}) (primitive.ObjectID, error) {
	result, err := c.Collection.InsertOne(ctx, document)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return result.InsertedID.(primitive.ObjectID), err
}

func (c *Collection) InsertMany(ctx context.Context, documents []interface{}) error {
	if _, err := c.Collection.InsertMany(ctx, documents); err != nil {
		return err
	}

	return nil
}

func (c *Collection) UpdateOne(ctx context.Context, filter F, update interface{}) error {
	if _, err := c.Collection.UpdateOne(ctx, filter, bson.M{"$set": update}); err != nil {
		return err
	}

	return nil
}

func (c *Collection) UpdateByID(ctx context.Context, id primitive.ObjectID, update interface{}) error {
	if _, err := c.Collection.UpdateByID(ctx, id, bson.M{"$set": update}); err != nil {
		return err
	}

	return nil
}

func (c *Collection) UpdateMany(ctx context.Context, filter F, update interface{}) error {
	if _, err := c.Collection.UpdateMany(ctx, filter, bson.M{"$set": update}); err != nil {
		return err
	}

	return nil
}

func (c *Collection) DeleteOne(ctx context.Context, filter F) error {
	if _, err := c.Collection.DeleteOne(ctx, filter); err != nil {
		return err
	}

	return nil
}

func (c *Collection) DeleteByID(ctx context.Context, id primitive.ObjectID) error {
	if _, err := c.Collection.DeleteOne(ctx, Key_ID.Eq(id)); err != nil {
		return err
	}

	return nil
}

func (c *Collection) DeleteMany(ctx context.Context, filter F) error {
	if _, err := c.Collection.DeleteMany(ctx, filter); err != nil {
		return err
	}

	return nil
}

func (c *Collection) FindOne(ctx context.Context, filter F, result interface{}) error {
	if err := c.Collection.FindOne(ctx, filter).Decode(result); err != nil {
		return err
	}

	return nil
}

func (c *Collection) FindByID(ctx context.Context, id primitive.ObjectID, result interface{}) error {
	if err := c.Collection.FindOne(ctx, Key_ID.Eq(id)).Decode(result); err != nil {
		return err
	}

	return nil
}

func (c *Collection) FindOneOrZero(ctx context.Context, filter F, result interface{}) error {
	if err := c.Collection.FindOne(ctx, filter).Decode(result); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil
		}

		return err
	}

	return nil
}

func (c *Collection) Find(ctx context.Context, filter F, results interface{}) error {
	cursor, err := c.Collection.Find(ctx, filter)
	if err != nil {
		return err
	}

	return cursor.All(ctx, results)
}

func (c *Collection) FindSort(ctx context.Context, filter F, sort S, results interface{}) error {
	cursor, err := c.Collection.Find(ctx, filter, options.Find().SetSort(sort))
	if err != nil {
		return err
	}

	return cursor.All(ctx, results)
}

func (c *Collection) FindSortLimit(ctx context.Context, filter F, sort S, limit int64, results interface{}) error {
	cursor, err := c.Collection.Find(ctx, filter, options.Find().SetSort(sort).SetLimit(limit))
	if err != nil {
		return err
	}

	return cursor.All(ctx, results)
}

func (c *Collection) FindSortLimitSkip(ctx context.Context, filter F, sort S, limit int64, skip int64, results interface{}) error {
	cursor, err := c.Collection.Find(ctx, filter, options.Find().SetSort(sort).SetLimit(limit).SetSkip(skip))
	if err != nil {
		return err
	}

	return cursor.All(ctx, results)
}

func (c *Collection) Aggregate(ctx context.Context, pipeline interface{}, results interface{}) error {
	cursor, err := c.Collection.Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}

	return cursor.All(ctx, results)
}

type F bson.D

func (f F) And(filter F) F {
	return F{
		bson.E{Key: "$and", Value: append([]F{f}, filter)},
	}
}

func (f F) Or(filter F) F {
	return F{
		bson.E{Key: "$or", Value: append([]F{f}, filter)},
	}
}

type U bson.M

func (u U) Set(key K, value interface{}) U {
	u[string(key)] = value
	return u
}

func Update() U {
	return U{}
}

type S bson.D
type Proj bson.D

type K string

const Key_ID K = "_id"

func Key(k string) K {
	return K(k)
}

func (k K) Eq(value interface{}) F {
	return F{
		bson.E{Key: string(k), Value: value},
	}
}

func (k K) Ne(value interface{}) F {
	return F{
		bson.E{Key: string(k), Value: F{{Key: "$ne", Value: value}}},
	}
}

func (k K) Gt(value interface{}) F {
	return F{
		bson.E{Key: string(k), Value: F{{Key: "$gt", Value: value}}},
	}
}

func (k K) Gte(value interface{}) F {
	return F{
		bson.E{Key: string(k), Value: F{{Key: "$gte", Value: value}}},
	}
}

func (k K) Lt(value interface{}) F {
	return F{
		bson.E{Key: string(k), Value: F{{Key: "$lt", Value: value}}},
	}
}

func (k K) Lte(value interface{}) F {
	return F{
		bson.E{Key: string(k), Value: F{{Key: "$lte", Value: value}}},
	}
}

func (k K) In(value interface{}) F {
	return F{
		bson.E{Key: string(k), Value: F{{Key: "$in", Value: value}}},
	}
}

func (k K) Nin(value interface{}) F {
	return F{
		bson.E{Key: string(k), Value: F{{Key: "$nin", Value: value}}},
	}
}

func (k K) Exists(value bool) F {
	return F{
		bson.E{Key: string(k), Value: F{{Key: "$exists", Value: value}}},
	}
}

func (k K) Regex(value string) F {
	return F{
		bson.E{Key: string(k), Value: F{{Key: "$regex", Value: value}}},
	}
}

func (k K) Asc() S {
	return S{
		bson.E{Key: string(k), Value: 1},
	}
}

func (k K) Desc() S {
	return S{
		bson.E{Key: string(k), Value: -1},
	}
}

func (k K) Proj() Proj {
	return Proj{
		bson.E{Key: string(k), Value: 1},
	}
}

func (k K) Set(v interface{}) U {
	return U{
		string(k): v,
	}
}
