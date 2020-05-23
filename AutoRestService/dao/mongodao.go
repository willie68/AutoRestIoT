package dao

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/willie68/AutoRestIoT/config"
	"github.com/willie68/AutoRestIoT/internal"
	"github.com/willie68/AutoRestIoT/internal/crypt"
	"github.com/willie68/AutoRestIoT/internal/slicesutils"
	"github.com/willie68/AutoRestIoT/logging"
	"github.com/willie68/AutoRestIoT/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const timeout = 1 * time.Minute
const attachmentsCollectionName = "attachments"
const usersCollectionName = "users"

// MongoDAO a mongodb based dao
type MongoDAO struct {
	initialised bool
	client      *mongo.Client
	mongoConfig config.MongoDB
	bucket      gridfs.Bucket
	database    mongo.Database
	ticker      time.Ticker
	done        chan bool
}

var log logging.ServiceLogger

//InitDAO initialise the mongodb connection, build up all collections and indexes
func (m *MongoDAO) InitDAO(MongoConfig config.MongoDB) {
	m.initialised = false
	m.mongoConfig = MongoConfig
	//	uri := fmt.Sprintf("mongodb://%s:%s@%s:%d", mongoConfig.Username, mongoConfig.Password, mongoConfig.Host, mongoConfig.Port)
	uri := fmt.Sprintf("mongodb://%s:%d", m.mongoConfig.Host, m.mongoConfig.Port)
	clientOptions := options.Client()
	clientOptions.ApplyURI(uri)
	clientOptions.Auth = &options.Credential{Username: m.mongoConfig.Username, Password: m.mongoConfig.Password, AuthSource: m.mongoConfig.AuthDB}
	var err error
	m.client, err = mongo.NewClient(clientOptions)
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	err = m.client.Connect(ctx)
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
	}
	m.database = *m.client.Database(m.mongoConfig.Database)

	myBucket, err := gridfs.NewBucket(&m.database, options.GridFSBucket().SetName(attachmentsCollectionName))
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
	}
	m.bucket = *myBucket

	m.initialised = true
}

//ProcessFiles adding a file to the storage, stream like
func (m *MongoDAO) ProcessFiles(RemoveCallback func(fileInfo model.FileInfo) bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cursor, err := m.bucket.Find(bson.M{}, &options.GridFSFindOptions{})
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return err
	}
	defer cursor.Close(ctx)
	count := 0
	for cursor.Next(ctx) {
		var file bson.M
		if err = cursor.Decode(&file); err != nil {
			log.Alertf("%v", err)
			continue
		}
		metadata := file["metadata"].(bson.M)
		info := model.FileInfo{}
		info.Filename = file["filename"].(string)
		info.Backend = metadata["backend"].(string)
		info.ID = file["_id"].(primitive.ObjectID).Hex()
		info.UploadDate = file["uploadDate"].(primitive.DateTime).Time()
		RemoveCallback(info)
		count++
	}
	return nil
}

//AddFile adding a file to the storage, stream like
func (m *MongoDAO) AddFile(backend string, filename string, reader io.Reader) (string, error) {
	uploadOpts := options.GridFSUpload().SetMetadata(bson.D{{Key: "backend", Value: backend}})

	fileID, err := m.bucket.UploadFromStream(filename, reader, uploadOpts)
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return "", err
	}
	log.Infof("Write file to DB was successful. File id: %s \n", fileID)
	id := fileID.Hex()
	return id, nil
}

//GetFilename getting the filename of an attachment from the database with the id
func (m *MongoDAO) GetFilename(backend string, fileid string) (string, error) {
	objectID, err := primitive.ObjectIDFromHex(fileid)
	if err != nil {
		log.Alertf("%v", err)
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cursor, err := m.bucket.Find(bson.M{"_id": objectID, "metadata.backend": backend})
	if err != nil {
		log.Alertf("%v", err)
		return "", err
	}
	defer cursor.Close(ctx)
	cursor.Next(ctx)
	var file bson.M
	var filename string
	if err = cursor.Decode(&file); err != nil {
		log.Alertf("%v", err)
		return "", err
	}
	filename = file["filename"].(string)
	return filename, nil
}

//GetFile getting a single file from the database with the id
func (m *MongoDAO) GetFile(backend string, fileid string, stream io.Writer) error {
	_, err := m.GetFilename(backend, fileid)
	if err != nil {
		log.Alertf("%v", err)
		return err
	}

	objectID, err := primitive.ObjectIDFromHex(fileid)
	if err != nil {
		log.Alertf("%v", err)
		return err
	}
	_, err = m.bucket.DownloadToStream(objectID, stream)
	if err != nil {
		log.Alertf("%v", err)
		return err
	}
	return nil
}

//DeleteFile getting a single from the database with the id
func (m *MongoDAO) DeleteFile(backend string, fileid string) error {
	_, err := m.GetFilename(backend, fileid)
	if err != nil {
		log.Alertf("%v", err)
		return err
	}

	objectID, err := primitive.ObjectIDFromHex(fileid)
	if err != nil {
		log.Alertf("%v", err)
		return err
	}
	err = m.bucket.Delete(objectID)
	if err != nil {
		log.Alertf("%v", err)
		return err
	}
	return nil
}

//GetUsers getting a list of users
func (m *MongoDAO) GetUsers() ([]model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	usersCollection := m.database.Collection(usersCollectionName)
	filter := bson.M{}
	cursor, err := usersCollection.Find(ctx, filter)
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return nil, err
	}
	defer cursor.Close(ctx)
	users := make([]model.User, 0)
	for cursor.Next(ctx) {
		var user model.User
		if err = cursor.Decode(&user); err != nil {
			log.Alertf("%v", err)
			return nil, err
		}
		users = append(users, user)
	}

	sort.Slice(users, func(i, j int) bool {
		return users[i].Name < users[j].Name
	})
	return users, nil
}

//GetUser getting the usermodel
func (m *MongoDAO) GetUser(username string) (model.User, bool) {
	username = strings.ToLower(username)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	usersCollection := m.database.Collection(usersCollectionName)
	var user model.User
	filter := bson.M{"name": username}
	err := usersCollection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return model.User{}, false
	}
	password := user.Password
	hash := BuildPasswordHash(password, user.Salt)
	user.Password = hash
	return user, true
}

// AddUser adding a new user to the system
func (m *MongoDAO) AddUser(user model.User) (model.User, error) {
	if user.Name == "" {
		return model.User{}, errors.New("username should not be empty")
	}
	user.Name = strings.ToLower(user.Name)
	_, ok := m.GetUser(user.Name)
	if ok {
		return model.User{}, errors.New("username already exists")
	}

	user.Salt, _ = crypt.GenerateRandomBytes(20)
	user.Password = BuildPasswordHash(user.Password, user.Salt)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	collection := m.database.Collection(usersCollectionName)
	_, err := collection.InsertOne(ctx, user)
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return model.User{}, err
	}
	return user, nil
}

// DeleteUser deletes one user from the system
func (m *MongoDAO) DeleteUser(username string) error {
	if username == "" {
		return errors.New("username should not be empty")
	}
	username = strings.ToLower(username)
	_, ok := m.GetUser(username)
	if !ok {
		return errors.New("username not exists")
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	collection := m.database.Collection(usersCollectionName)
	filter := bson.M{"name": username}
	_, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return err
	}
	return nil
}

// ChangePWD changes the apssword of a single user
func (m *MongoDAO) ChangePWD(username string, newpassword string) (model.User, error) {
	if username == "" {
		return model.User{}, errors.New("username should not be empty")
	}
	username = strings.ToLower(username)
	userModel, ok := m.GetUser(username)
	if !ok {
		return model.User{}, errors.New("username not registered")
	}

	newsalt, _ := crypt.GenerateRandomBytes(20)
	newpassword = BuildPasswordHash(newpassword, newsalt)
	userModel.Password = newpassword
	userModel.Salt = newsalt

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	collection := m.database.Collection(usersCollectionName)
	filter := bson.M{"name": username}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "password", Value: newpassword}, {Key: "salt", Value: newsalt}}}}
	result := collection.FindOneAndUpdate(ctx, filter, update)
	if result.Err() != nil {
		fmt.Printf("error: %s\n", result.Err().Error())
		return model.User{}, result.Err()
	}
	return userModel, nil
}

//CreateModel creating a new model
func (m *MongoDAO) CreateModel(route model.Route, data model.JSONMap) (string, error) {
	collectionName := route.GetRouteName()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	collection := m.database.Collection(collectionName)
	result, err := collection.InsertOne(ctx, data)
	if err != nil {
		switch v := err.(type) {
		case mongo.WriteException:
			if v.WriteErrors[0].Code == 11000 {
				return "", ErrUniqueIndexError
			}
		}
		fmt.Printf("error: %s\n", err.Error())
		return "", err
	}
	switch v := result.InsertedID.(type) {
	case primitive.ObjectID:
		return v.Hex(), nil
	}
	return "", ErrUnknownError
}

//CreateModels creates a bunch of models
func (m *MongoDAO) CreateModels(route model.Route, datas []model.JSONMap) ([]string, error) {
	collectionName := route.GetRouteName()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	collection := m.database.Collection(collectionName)
	models := make([]interface{}, 0)
	for _, data := range datas {
		models = append(models, data)
	}
	result, err := collection.InsertMany(ctx, models, &options.InsertManyOptions{})
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return nil, err
	}
	modelids := make([]string, 0)
	for _, id := range result.InsertedIDs {
		switch v := id.(type) {
		case primitive.ObjectID:
			modelids = append(modelids, v.Hex())
		}
	}
	return modelids, nil
}

//GetModel getting requested model from the storage
func (m *MongoDAO) GetModel(route model.Route) (model.JSONMap, error) {
	collectionName := route.GetRouteName()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	collection := m.database.Collection(collectionName)
	objectID, _ := primitive.ObjectIDFromHex(route.Identity)
	result := collection.FindOne(ctx, bson.M{"_id": objectID})
	err := result.Err()
	if err == mongo.ErrNoDocuments {
		log.Alertf("%v", err)
		return nil, ErrNoDocument
	}
	if err != nil {
		log.Alertf("%v", err)
		return nil, err
	}
	var bemodel model.JSONMap
	if err := result.Decode(&bemodel); err != nil {
		log.Alertf("%v", err)
		return nil, err
	}
	//		bemodel[internal.AttributeID] = bemodel[internal.AttributeID].(primitive.ObjectID).Hex()
	bemodel, _ = m.convertModel(bemodel)
	return bemodel, nil
}

//CountModel counting all medelsin this collection
func (m *MongoDAO) CountModel(route model.Route) (int, error) {
	collectionName := route.GetRouteName()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	collection := m.database.Collection(collectionName)
	n, err := collection.CountDocuments(ctx, bson.M{}, &options.CountOptions{})
	if err == mongo.ErrNoDocuments {
		log.Alertf("%v", err)
		return 0, ErrNoDocument
	}
	if err != nil {
		log.Alertf("%v", err)
		return 0, err
	}
	return int(n), nil
}

//QueryModel query for the right models
func (m *MongoDAO) QueryModel(route model.Route, query string, offset int, limit int) (int, []model.JSONMap, error) {
	collectionName := route.GetRouteName()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	collection := m.database.Collection(collectionName)

	var queryM map[string]interface{}
	if query == "" {
		queryM = make(map[string]interface{})
	} else {
		err := json.Unmarshal([]byte(query), &queryM)
		if err != nil {
			log.Alertf("%v", err)
			return 0, nil, err
		}
	}

	queryDoc := bson.M{}
	for k, v := range queryM {
		if k == "$fulltext" {
			queryDoc["$text"] = bson.M{"$search": v}
		} else {
			switch v := v.(type) {
			//			case float64:
			//			case int:
			//			case bool:
			case string:
				queryDoc[k] = bson.M{"$regex": v}
			}
			//queryDoc[k] = v
		}
	}
	//data, _ := json.Marshal(queryDoc)
	//log.Infof("mongoquery: %s", string(data))
	n, err := collection.CountDocuments(ctx, queryDoc, &options.CountOptions{Collation: &options.Collation{Locale: "en", Strength: 2}})
	if err != nil {
		log.Alertf("%v", err)
		return 0, nil, err
	}
	cursor, err := collection.Find(ctx, queryDoc, &options.FindOptions{Collation: &options.Collation{Locale: "en", Strength: 2}})
	if err != nil {
		log.Alertf("%v", err)
		return 0, nil, err
	}
	defer cursor.Close(ctx)
	models := make([]model.JSONMap, 0)
	count := 0
	docs := 0
	for cursor.Next(ctx) {
		if count >= offset {
			if docs < limit {
				var model model.JSONMap
				if err = cursor.Decode(&model); err != nil {
					log.Alertf("%v", err)
					return 0, nil, err
				}
				models = append(models, model)
				docs++
			} else {
				break
			}
		}
		count++
	}
	return int(n), models, nil
}

//UpdateModel updateing an existing datamodel in the mongo db
func (m *MongoDAO) UpdateModel(route model.Route, data model.JSONMap) (model.JSONMap, error) {
	collectionName := route.GetRouteName()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	collection := m.database.Collection(collectionName)
	objectID, _ := primitive.ObjectIDFromHex(route.Identity)
	delete(data, internal.AttributeID)

	filter := bson.M{internal.AttributeID: objectID}
	updateResult, err := collection.ReplaceOne(ctx, filter, data)
	if err != nil {
		return nil, err
	}
	if updateResult.ModifiedCount == 0 {
		return nil, ErrUnknownError
	}
	newModel, err := m.GetModel(route)
	if err != nil {
		return nil, err
	}
	return newModel, nil
}

//DeleteModel deleting the requested model from the storage
func (m *MongoDAO) DeleteModel(route model.Route) error {
	collectionName := route.GetRouteName()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	collection := m.database.Collection(collectionName)
	objectID, _ := primitive.ObjectIDFromHex(route.Identity)

	filter := bson.M{internal.AttributeID: objectID}
	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if result.DeletedCount != 1 {
		return ErrUnknownError
	}
	return nil
}

// GetIndexNames getting a list of index names
func (m *MongoDAO) GetIndexNames(route model.Route) ([]string, error) {
	collection := m.database.Collection(route.GetRouteName())
	indexView := collection.Indexes()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cursor, err := indexView.List(ctx)
	if err != nil {
		log.Alertf("%v", err)
		return nil, err
	}
	defer cursor.Close(ctx)
	myIndexes := make([]string, 0)
	for cursor.Next(ctx) {
		var index bson.M
		if err = cursor.Decode(&index); err != nil {
			log.Alertf("%v", err)
			return nil, err
		}
		name := index["name"].(string)
		if !strings.HasPrefix(name, "_") {
			if name == "$text" {
				name = FulltextIndexName
			}
			myIndexes = append(myIndexes, name)
		}
	}

	return myIndexes, nil
}

// DeleteIndex delete one search index
func (m *MongoDAO) DeleteIndex(route model.Route, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	collection := m.database.Collection(route.GetRouteName())
	_, err := collection.Indexes().DropOne(ctx, name)
	if err != nil {
		log.Alertf("%v", err)
		return err
	}
	return nil
}

//UpdateIndex create or update an index
func (m *MongoDAO) UpdateIndex(route model.Route, index model.Index) error {
	myIndexes, err := m.GetIndexNames(route)
	if err != nil {
		log.Alertf("%v", err)
		return err
	}

	collection := m.database.Collection(route.GetRouteName())
	indexView := collection.Indexes()

	if !slicesutils.Contains(myIndexes, index.Name) {
		var indexmodel mongo.IndexModel
		if index.Name == FulltextIndexName {
			keys := bson.D{}
			for _, field := range index.Fields {
				//TODO here must be implemented the right field type
				keys = append(keys, primitive.E{
					Key:   field,
					Value: "text",
				})
			}
			indexmodel = mongo.IndexModel{
				Keys:    keys,
				Options: options.Index().SetName("$text"),
			}
		} else {
			keys := bson.D{}
			for _, field := range index.Fields {
				keys = append(keys, primitive.E{
					Key:   field,
					Value: 1,
				})
			}
			// TODO here must be implemented the right language
			idxOptions := options.Index().SetName(index.Name).SetCollation(&options.Collation{Locale: "en", Strength: 2})
			if index.Unique {
				idxOptions = idxOptions.SetUnique(true)
			}
			indexmodel = mongo.IndexModel{
				Keys:    keys,
				Options: idxOptions,
			}
		}

		// Specify the MaxTime option to limit the amount of time the operation can run on the server
		opts := options.CreateIndexes().SetMaxTime(2 * time.Second)
		name, err := indexView.CreateOne(context.TODO(), indexmodel, opts)
		if err != nil {
			log.Alertf("%v", err)
			return err
		}
		log.Infof("Index %s for route %s created.", name, route.GetRouteName())
	}
	return nil
}

// Ping pinging the mongoDao
func (m *MongoDAO) Ping() error {
	if !m.initialised {
		return errors.New("mongo client not initialised")
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return m.database.Client().Ping(ctx, nil)
}

// DeleteBackend dropping all data from the backend
func (m *MongoDAO) DeleteBackend(backend string) error {
	if backend == attachmentsCollectionName || backend == usersCollectionName {
		return errors.New("wrong backend name")
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	collectionNames, err := m.database.ListCollectionNames(ctx, bson.D{}, &options.ListCollectionsOptions{})
	if err != nil {
		log.Alertf("%v", err)
		return err
	}
	for _, name := range collectionNames {
		if strings.HasPrefix(name, backend+".") {
			collection := m.database.Collection(name)
			err = collection.Drop(ctx)
			if err != nil {
				log.Alertf("%v", err)
				return err
			}
		}
	}

	filter := bson.M{"metadata.backend": backend}
	cursor, err := m.bucket.Find(filter, &options.GridFSFindOptions{})
	if err != nil {
		log.Alertf("%v", err)
		return err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var file bson.M
		if err = cursor.Decode(&file); err != nil {
			log.Alertf("%v", err)
			return err
		}
		if err = m.bucket.Delete(file["_id"]); err != nil {
			log.Alertf("%v", err)
			return err
		}
	}
	return nil
}

// DropAll dropping all data from the database
func (m *MongoDAO) DropAll() {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	collectionNames, err := m.database.ListCollectionNames(ctx, bson.D{}, &options.ListCollectionsOptions{})
	if err != nil {
		log.Alertf("%v", err)
		return
	}
	for _, name := range collectionNames {
		if name != usersCollectionName {
			collection := m.database.Collection(name)
			err = collection.Drop(ctx)
			if err != nil {
				log.Alertf("%v", err)
				return
			}
		}
	}
}

// Stop stopping the mongodao
func (m *MongoDAO) Stop() {
	m.ticker.Stop()
	m.done <- true
}

func (m *MongoDAO) convertModel(srcModel model.JSONMap) (model.JSONMap, error) {
	dstModel := srcModel
	for k, v := range srcModel {
		dstModel[k] = m.convertValue(v)
	}
	return dstModel, nil
}

func (m *MongoDAO) convertValue(value interface{}) interface{} {
	switch v := value.(type) {
	case primitive.ObjectID:
		return v.Hex()
	case primitive.A:
		items := make([]interface{}, 0)
		for _, itemValue := range v {
			items = append(items, m.convertValue(itemValue))
		}
		return items
	}
	return value
}
