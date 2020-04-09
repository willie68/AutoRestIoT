package dao

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/willie68/AutoRestIoT/config"
	"github.com/willie68/AutoRestIoT/internal"
	"github.com/willie68/AutoRestIoT/internal/slicesutils"
	"github.com/willie68/AutoRestIoT/logging"
	"github.com/willie68/AutoRestIoT/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// time to reload all users
const userReloadPeriod = 1 * time.Hour
const timeout = 1 * time.Minute
const attachmentsCollectionName = "attachments"
const usersCollectionName = "users"
const fulltextIndexName = "$fulltext"

// MongoDAO a mongodb based dao
type MongoDAO struct {
	initialised bool
	client      *mongo.Client
	mongoConfig config.MongoDB
	bucket      gridfs.Bucket
	database    mongo.Database
	users       map[string]string
	ticker      time.Ticker
	done        chan bool
}

var log logging.ServiceLogger

// InitDAO initialise the mongodb connection, build up all collections and indexes
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

	m.initUsers()

	m.initialised = true
}

func (m *MongoDAO) initUsers() {
	m.reloadUsers()

	go func() {
		background := time.NewTicker(userReloadPeriod)
		for _ = range background.C {
			m.reloadUsers()
		}
	}()
}

func (m *MongoDAO) reloadUsers() {
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	usersCollection := m.database.Collection(usersCollectionName)
	cursor, err := usersCollection.Find(ctx, bson.M{})
	if err != nil {
		log.Alertf("%v", err)
	}
	defer cursor.Close(ctx)
	localUsers := make(map[string]string)
	for cursor.Next(ctx) {
		var user bson.M
		if err = cursor.Decode(&user); err != nil {
			log.Alertf("%v", err)
		} else {
			username := strings.ToLower(user["name"].(string))
			password := user["password"].(string)
			localUsers[username] = BuildPasswordHash(password)
		}
	}
	m.users = localUsers
	if len(m.users) == 0 {
		admin := model.User{
			Name:     "admin",
			Password: "admin",
			Admin:    true,
			Roles:    []string{"admin"},
		}
		m.AddUser(admin)
		editor := model.User{
			Name:     "editor",
			Password: "editor",
			Admin:    false,
			Guest:    false,
			Roles:    []string{"edit"},
		}
		m.AddUser(editor)
		guest := model.User{
			Name:     "guest",
			Password: "guest",
			Admin:    false,
			Guest:    true,
			Roles:    []string{"read"},
		}
		m.AddUser(guest)
	}
}

// AddFile adding a file to the storage, stream like
func (m *MongoDAO) AddFile(backend string, filename string, reader io.Reader) (string, error) {
	uploadOpts := options.GridFSUpload().SetMetadata(bson.D{{"backend", backend}})

	fileID, err := m.bucket.UploadFromStream(filename, reader, uploadOpts)
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return "", err
	}
	log.Infof("Write file to DB was successful. File id: %s \n", fileID)
	id := fileID.Hex()
	return id, nil
}

//GetFile getting a single from the database with the id
func (m *MongoDAO) GetFilename(backend string, fileid string) (string, error) {
	objectID, err := primitive.ObjectIDFromHex(fileid)
	if err != nil {
		log.Alertf("%v", err)
		return "", err
	}
	ctx, _ := context.WithTimeout(context.Background(), timeout)
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
	} else {
		filename = file["filename"].(string)
	}
	return filename, nil
}

//GetFile getting a single from the database with the id
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

// CheckUser checking username and password... returns true if the user is active and the password for this user is correct
func (m *MongoDAO) CheckUser(username string, password string) bool {
	username = strings.ToLower(username)
	pwd, ok := m.users[username]
	if ok {
		if pwd == password {
			return true
		} else {
			user, ok := m.GetUser(username)
			if ok {
				if user.Password == password {
					return true
				}
			}
		}
	}

	if !ok {
		user, ok := m.GetUser(username)
		if ok {
			if user.Password == password {
				return true
			}
		}
	}

	return false
}

//UserInRoles is a user in the given role
func (m *MongoDAO) UserInRoles(username string, roles []string) bool {
	user, ok := m.GetUser(username)
	if !ok {
		return false
	}

	for _, role := range roles {
		if slicesutils.Contains(user.Roles, role) {
			return true
		}
	}
	return false
}

// GetUsers getting a list of users
func (m *MongoDAO) GetUsers() ([]model.User, error) {
	ctx, _ := context.WithTimeout(context.Background(), timeout)
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
		} else {
			user.Password = ""
			users = append(users, user)
		}
	}
	return users, nil
}

// GetUser getting the usermodel
func (m *MongoDAO) GetUser(username string) (model.User, bool) {
	username = strings.ToLower(username)
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	usersCollection := m.database.Collection(usersCollectionName)
	var user model.User
	filter := bson.M{"name": username}
	err := usersCollection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return model.User{}, false
	}
	password := user.Password
	hash := BuildPasswordHash(password)
	m.users[username] = hash
	return user, true
}

// AddUser adding a new user to the system
func (m *MongoDAO) AddUser(user model.User) error {
	if user.Name == "" {
		return errors.New("username should not be empty")
	}
	user.Name = strings.ToLower(user.Name)
	_, ok := m.users[user.Name]
	if ok {
		return errors.New("username already exists")
	}

	user.Password = BuildPasswordHash(user.Password)

	ctx, _ := context.WithTimeout(context.Background(), timeout)
	collection := m.database.Collection(usersCollectionName)
	_, err := collection.InsertOne(ctx, user)
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return err
	}
	m.users[user.Name] = user.Password
	return nil
}

// DeleteUser deletes one user from the system
func (m *MongoDAO) DeleteUser(username string) error {
	if username == "" {
		return errors.New("username should not be empty")
	}
	username = strings.ToLower(username)
	_, ok := m.users[username]
	if !ok {
		return errors.New("username not exists")
	}

	ctx, _ := context.WithTimeout(context.Background(), timeout)
	collection := m.database.Collection(usersCollectionName)
	filter := bson.M{"name": username}
	_, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return err
	}
	delete(m.users, username)
	return nil
}

// ChangePWD changes the apssword of a single user
func (m *MongoDAO) ChangePWD(username string, newpassword string, oldpassword string) error {
	if username == "" {
		return errors.New("username should not be empty")
	}
	username = strings.ToLower(username)
	pwd, ok := m.users[username]
	if !ok {
		return errors.New("username not registered")
	}

	newpassword = BuildPasswordHash(newpassword)
	oldpassword = BuildPasswordHash(oldpassword)
	if pwd != oldpassword {
		return errors.New("actual password incorrect")
	}

	ctx, _ := context.WithTimeout(context.Background(), timeout)
	collection := m.database.Collection(usersCollectionName)
	filter := bson.M{"name": username}
	update := bson.D{{"$set", bson.D{{"password", newpassword}}}}
	result := collection.FindOneAndUpdate(ctx, filter, update)
	if result.Err() != nil {
		fmt.Printf("error: %s\n", result.Err().Error())
		return result.Err()
	}
	m.users[username] = newpassword
	return nil
}

func (m *MongoDAO) CreateModel(route model.Route, data model.JsonMap) (string, error) {
	collectionName := route.GetRouteName()
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	collection := m.database.Collection(collectionName)
	result, err := collection.InsertOne(ctx, data)
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return "", err
	}
	switch v := result.InsertedID.(type) {
	case primitive.ObjectID:
		return v.Hex(), nil
	}
	return "", ErrUnknownError
}

func (m *MongoDAO) GetModel(route model.Route) (model.JsonMap, error) {
	collectionName := route.GetRouteName()
	ctx, _ := context.WithTimeout(context.Background(), timeout)
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
	var bemodel model.JsonMap
	if err := result.Decode(&bemodel); err != nil {
		log.Alertf("%v", err)
		return nil, err
	} else {
		//		bemodel[internal.AttributeID] = bemodel[internal.AttributeID].(primitive.ObjectID).Hex()
		bemodel, _ = m.convertModel(bemodel)
		return bemodel, nil
	}
}

func (m *MongoDAO) Query(route model.Route, query string, offset int, limit int) (int, []model.JsonMap, error) {
	return 0, nil, ErrNotImplemented
}

//UpdateModel updateing an existing datamodel in the mongo db
func (m *MongoDAO) UpdateModel(route model.Route, data model.JsonMap) (model.JsonMap, error) {
	collectionName := route.GetRouteName()
	ctx, _ := context.WithTimeout(context.Background(), timeout)
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

func (m *MongoDAO) DeleteModel(route model.Route) error {
	collectionName := route.GetRouteName()
	ctx, _ := context.WithTimeout(context.Background(), timeout)
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
	ctx, _ := context.WithTimeout(context.Background(), timeout)
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
				name = fulltextIndexName
			}
			myIndexes = append(myIndexes, name)
		}
	}

	return myIndexes, nil
}

// DeleteIndex delete one search index
func (m *MongoDAO) DeleteIndex(route model.Route, name string) error {
	ctx, _ := context.WithTimeout(context.Background(), timeout)
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
		if index.Name == fulltextIndexName {
			keys := bson.D{}
			for _, field := range index.Fields {
				//TODO here must be im plemented the right field type
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
			// TODO here must be im pleneted the right language
			indexmodel = mongo.IndexModel{
				Keys:    keys,
				Options: options.Index().SetName(index.Name).SetCollation(&options.Collation{Locale: "en", Strength: 2}),
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
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	return m.database.Client().Ping(ctx, nil)
}

// DeleteBackend dropping all data from the backend
func (m *MongoDAO) DeleteBackend(backend string) error {
	if backend == attachmentsCollectionName || backend == usersCollectionName {
		return errors.New("wrong backend name.")
	}
	ctx, _ := context.WithTimeout(context.Background(), timeout)
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
		} else {
			if err = m.bucket.Delete(file["_id"]); err != nil {
				log.Alertf("%v", err)
				return err
			}
		}

	}

	return nil
}

// DropAll dropping all data from the database
func (m *MongoDAO) DropAll() {
	ctx, _ := context.WithTimeout(context.Background(), timeout)
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

func (m *MongoDAO) convertModel(srcModel model.JsonMap) (model.JsonMap, error) {
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
