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

	filter F
	sort   S
	proj   Proj
	skip   int64
	limit  int64
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

		filter: F{},
		sort:   S{},
		proj:   Proj{},
		skip:   -1,
		limit:  -1,
	}
}

func (c *Collection) Sort(sorts ...S) *Collection {
	if len(sorts) == 0 {
		return c
	}

	for _, sort := range sorts {
		c.sort = append(c.sort, sort...)
	}

	return c
}

const (
	NoUse int64 = -1
)

func (c *Collection) Skip(skip int64) *Collection {
	c.skip = skip
	return c
}

func (c *Collection) Limit(limit int64) *Collection {
	c.limit = limit
	return c
}

func (c *Collection) Projection(proj ...Proj) *Collection {
	if len(proj) == 0 {
		return c
	}

	for _, p := range proj {
		c.proj = append(c.proj, p...)
	}

	return c
}

func (c *Collection) Where(filters ...F) *Collection {
	if len(filters) == 0 {
		c.filter = F{}
		return c
	}

	c.filter = F{bson.E{Key: "$and", Value: filters}}

	return c
}

func (c *Collection) WhereAnd(filter F) *Collection {
	if len(c.filter) == 0 {
		c.filter = filter

		return c
	}

	c.filter = F{bson.E{Key: "$and", Value: append([]F{c.filter}, filter)}}

	return c
}

func (c *Collection) WhereOr(filter F) *Collection {
	if len(c.filter) == 0 {
		c.filter = filter

		return c
	}

	c.filter = F{bson.E{Key: "$or", Value: append([]F{c.filter}, filter)}}

	return c
}

func (c *Collection) InsertOne(ctx context.Context, document any) (primitive.ObjectID, error) {
	result, err := c.Collection.InsertOne(ctx, document)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return result.InsertedID.(primitive.ObjectID), err
}

func (c *Collection) InsertMany(ctx context.Context, documents []any) error {
	if _, err := c.Collection.InsertMany(ctx, documents); err != nil {
		return err
	}

	return nil
}

func (c *Collection) UpdateOne(ctx context.Context, update any) error {
	if _, err := c.Collection.UpdateOne(ctx, c.filter, bson.M{"$set": update}); err != nil {
		return err
	}

	return nil
}

func (c *Collection) UpdateMany(ctx context.Context, update any) error {
	if _, err := c.Collection.UpdateMany(ctx, c.filter, bson.M{"$set": update}); err != nil {
		return err
	}

	return nil
}

func (c *Collection) Upsert(ctx context.Context, update any) error {
	if _, err := c.Collection.UpdateOne(ctx, c.filter, bson.M{"$set": update}, options.Update().SetUpsert(true)); err != nil {
		return err
	}

	return nil
}

func (c *Collection) DeleteOne(ctx context.Context) error {
	if _, err := c.Collection.DeleteOne(ctx, c.filter); err != nil {
		return err
	}

	return nil
}

func (c *Collection) DeleteMany(ctx context.Context) error {
	if _, err := c.Collection.DeleteMany(ctx, c.filter); err != nil {
		return err
	}

	return nil
}

func (c *Collection) FindOne(ctx context.Context, result any) error {
	opts := options.FindOne()
	if len(c.sort) > 0 {
		opts.SetSort(c.sort)
	}
	if c.skip > 0 {
		opts.SetSkip(c.skip)
	}
	if len(c.proj) > 0 {
		opts.SetProjection(c.proj)
	}

	if err := c.Collection.FindOne(ctx, c.filter, opts).Decode(result); err != nil {
		return err
	}

	return nil
}

func (c *Collection) FindOneOrZero(ctx context.Context, result any) error {
	opts := options.FindOne()
	if len(c.sort) > 0 {
		opts.SetSort(c.sort)
	}
	if c.skip > 0 {
		opts.SetSkip(c.skip)
	}
	if len(c.proj) > 0 {
		opts.SetProjection(c.proj)
	}

	if err := c.Collection.FindOne(ctx, c.filter, opts).Decode(result); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil
		}

		return err
	}

	return nil
}

func (c *Collection) Find(ctx context.Context, results any) error {
	opts := options.Find()
	if len(c.sort) > 0 {
		opts.SetSort(c.sort)
	}
	if c.limit > 0 {
		opts.SetLimit(c.limit)
	}
	if c.skip > 0 {
		opts.SetSkip(c.skip)
	}
	if len(c.proj) > 0 {
		opts.SetProjection(c.proj)
	}

	cursor, err := c.Collection.Find(ctx, c.filter, opts)
	if err != nil {
		return err
	}

	return cursor.All(ctx, results)
}

func (c *Collection) Count(ctx context.Context) (int64, error) {
	return c.Collection.CountDocuments(ctx, c.filter)
}

func (c *Collection) Aggregate(ctx context.Context, pipeline any, results any) error {
	cursor, err := c.Collection.Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}

	return cursor.All(ctx, results)
}

type F bson.D

func (f F) And(filter F) F {
	if len(f) == 0 {
		return filter
	}

	if len(filter) == 0 {
		return f
	}

	return F{
		bson.E{Key: "$and", Value: append([]F{f}, filter)},
	}
}

func (f F) Or(filter F) F {
	if len(f) == 0 {
		return filter
	}

	if len(filter) == 0 {
		return f
	}

	return F{
		bson.E{Key: "$or", Value: append([]F{f}, filter)},
	}
}

type U bson.M

func (u U) Set(key K, value any) U {
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

func (k K) Eq(value any) F {
	return F{
		bson.E{Key: string(k), Value: value},
	}
}

func (k K) Ne(value any) F {
	return F{
		bson.E{Key: string(k), Value: F{{Key: "$ne", Value: value}}},
	}
}

func (k K) Gt(value any) F {
	return F{
		bson.E{Key: string(k), Value: F{{Key: "$gt", Value: value}}},
	}
}

func (k K) Gte(value any) F {
	return F{
		bson.E{Key: string(k), Value: F{{Key: "$gte", Value: value}}},
	}
}

func (k K) Lt(value any) F {
	return F{
		bson.E{Key: string(k), Value: F{{Key: "$lt", Value: value}}},
	}
}

func (k K) Lte(value any) F {
	return F{
		bson.E{Key: string(k), Value: F{{Key: "$lte", Value: value}}},
	}
}

func (k K) In(value any) F {
	return F{
		bson.E{Key: string(k), Value: F{{Key: "$in", Value: value}}},
	}
}

func (k K) Nin(value any) F {
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

func (k K) NotProj() Proj {
	return Proj{
		bson.E{Key: string(k), Value: 0},
	}
}

func (k K) Set(v any) U {
	return U{
		string(k): v,
	}
}
