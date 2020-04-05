package dao

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/willie68/AutoRestIoT/config"
	slicesutils "github.com/willie68/AutoRestIoT/internal"
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
const schematicsCollectionName = "schematics"
const tagsCollectionName = "tags"
const manufacturersCollectionName = "manufacturers"
const usersCollectionName = "users"
const effectsCollectionName = "effects"
const effectTypesCollectionName = "effectTypes"

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

	/*
		m.initIndexSchematics()
		m.initIndexTags()
		m.initIndexManufacturers()
		m.initIndexEffectTypes()
		m.initIndexEffects()

		m.tags = model.NewTags()
		m.manufacturers = model.NewManufacturers()
		m.users = make(map[string]string)
		m.initTags()
		m.initManufacturers()
	*/
	m.initUsers()

	m.initialised = true
}

/*
func (m *MongoDAO) initIndexSchematics() {
	collection := m.database.Collection(schematicsCollectionName)
	indexView := collection.Indexes()
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	cursor, err := indexView.List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)
	myIndexes := make([]string, 0)
	for cursor.Next(ctx) {
		var index bson.M
		if err = cursor.Decode(&index); err != nil {
			log.Fatal(err)
		}
		myIndexes = append(myIndexes, index["name"].(string))
	}

	if !slicesutils.Contains(myIndexes, "manufacturer") {
		ctx, _ = context.WithTimeout(context.Background(), timeout)
		models := []mongo.IndexModel{
			{
				Keys:    bson.D{{"manufacturer", 1}},
				Options: options.Index().SetName("manufacturer").SetCollation(&options.Collation{Locale: "en", Strength: 2}),
			},
			{
				Keys:    bson.D{{"model", 1}},
				Options: options.Index().SetName("model").SetCollation(&options.Collation{Locale: "en", Strength: 2}),
			},
			{
				Keys:    bson.D{{"tags", 1}},
				Options: options.Index().SetName("tags").SetCollation(&options.Collation{Locale: "en", Strength: 2}),
			},
			{
				Keys:    bson.D{{"subtitle", 1}},
				Options: options.Index().SetName("subtitle").SetCollation(&options.Collation{Locale: "en", Strength: 2}),
			},
			{
				Keys:    bson.D{{"manufacturer", "text"}, {"model", "text"}, {"tags", "text"}, {"subtitle", "text"}, {"description", "text"}, {"owner", "text"}},
				Options: options.Index().SetName("$text"),
			},
		}

		// Specify the MaxTime option to limit the amount of time the operation can run on the server
		opts := options.CreateIndexes().SetMaxTime(2 * time.Second)
		names, err := indexView.CreateMany(context.TODO(), models, opts)
		if err != nil {
			log.Fatal(err)
		}
		log.Print("create indexes:")
		for _, name := range names {
			log.Println(name)
		}
	}
}
*/

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
		log.Fatal(err)
	}
	defer cursor.Close(ctx)
	localUsers := make(map[string]string)
	for cursor.Next(ctx) {
		var user bson.M
		if err = cursor.Decode(&user); err != nil {
			log.Fatal(err)
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
func (m *MongoDAO) AddFile(filename string, reader io.Reader) (string, error) {
	uploadOpts := options.GridFSUpload().SetMetadata(bson.D{{"tag", "tag"}})

	fileID, err := m.bucket.UploadFromStream(filename, reader, uploadOpts)
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return "", err
	}
	log.Printf("Write file to DB was successful. File id: %s \n", fileID)
	id := fileID.Hex()
	return id, nil
}

//GetFile getting a single from the database with the id
func (m *MongoDAO) GetFilename(fileid string) (string, error) {
	objectID, err := primitive.ObjectIDFromHex(fileid)
	if err != nil {
		log.Print(err)
		return "", err
	}
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	cursor, err := m.bucket.Find(bson.M{"_id": objectID})
	if err != nil {
		log.Print(err)
		return "", err
	}
	defer cursor.Close(ctx)
	cursor.Next(ctx)
	var file bson.M
	var filename string
	if err = cursor.Decode(&file); err != nil {
		log.Print(err)
		return "", err
	} else {
		filename = file["filename"].(string)
	}
	return filename, nil
}

//GetFile getting a single from the database with the id
func (m *MongoDAO) GetFile(fileid string, stream io.Writer) error {
	objectID, err := primitive.ObjectIDFromHex(fileid)
	if err != nil {
		log.Print(err)
		return err
	}
	_, err = m.bucket.DownloadToStream(objectID, stream)
	if err != nil {
		log.Print(err)
		return err
	}
	return nil
}

/*
// CreateSchematic creating a new schematic in the database
func (m *MongoDAO) CreateSchematic(schematic model.Schematic) (string, error) {

	for _, tag := range schematic.Tags {
		if !m.tags.Contains(tag) {
			m.CreateTag(tag)
		}
	}

	if !m.manufacturers.Contains(schematic.Manufacturer) {
		m.CreateManufacturer(schematic.Manufacturer)
	}

	ctx, _ := context.WithTimeout(context.Background(), timeout)
	collection := m.database.Collection(schematicsCollectionName)
	result, err := collection.InsertOne(ctx, schematic)
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return "", err
	}
	filter := bson.M{"_id": result.InsertedID}
	err = collection.FindOne(ctx, filter).Decode(&schematic)
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return "", err
	}
	switch v := result.InsertedID.(type) {
	case primitive.ObjectID:
		return v.Hex(), nil
	}
	return "", nil
}

// UpdateSchematic creating a new schematic in the database
func (m *MongoDAO) UpdateSchematic(schematic model.Schematic) (string, error) {

	for _, tag := range schematic.Tags {
		if !m.tags.Contains(tag) {
			m.CreateTag(tag)
		}
	}

	if !m.manufacturers.Contains(schematic.Manufacturer) {
		m.CreateManufacturer(schematic.Manufacturer)
	}

	ctx, _ := context.WithTimeout(context.Background(), timeout)
	collection := m.database.Collection(schematicsCollectionName)
	filter := bson.M{"_id": schematic.ID}
	updateDoc := bson.D{{"$set", schematic}}
	result, err := collection.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return "", err
	}
	if result.ModifiedCount != 1 {
		return "", errors.New("can't update document.")
	}
	err = collection.FindOne(ctx, filter).Decode(&schematic)
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return "", err
	}
	return schematic.ID.Hex(), nil
}

// GetSchematic getting a sdingle schematic
func (m *MongoDAO) GetSchematic(schematicID string) (model.Schematic, error) {
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	schematicCollection := m.database.Collection(schematicsCollectionName)
	objectID, _ := primitive.ObjectIDFromHex(schematicID)
	result := schematicCollection.FindOne(ctx, bson.M{"_id": objectID})
	err := result.Err()
	if err == mongo.ErrNoDocuments {
		log.Print(err)
		return model.Schematic{}, ErrNoDocument
	}
	if err != nil {
		log.Print(err)
		return model.Schematic{}, err
	}
	var schematic model.Schematic
	if err := result.Decode(&schematic); err != nil {
		log.Print(err)
		return model.Schematic{}, err
	} else {
		return schematic, nil
	}
}

// DeleteSchematic getting a sdingle schematic
func (m *MongoDAO) DeleteSchematic(schematicID string) error {
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	schematicCollection := m.database.Collection(schematicsCollectionName)
	objectID, _ := primitive.ObjectIDFromHex(schematicID)
	result, err := schematicCollection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		log.Print(err)
		return err
	} else {
		if result.DeletedCount > 0 {
			return nil
		}
		return ErrNoDocument
	}
}
*/
/*
// GetSchematics getting a sdingle schematic
func (m *MongoDAO) GetSchematics(query string, offset int, limit int, owner string) (int64, []model.Schematic, error) {
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	schematicCollection := m.database.Collection(schematicsCollectionName)
	var queryM map[string]interface{}
	err := json.Unmarshal([]byte(query), &queryM)
	if err != nil {
		log.Print(err)
		return 0, nil, err
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
	data, _ := json.Marshal(queryDoc)
	log.Printf("mongoquery: %s\n", string(data))
	n, err := schematicCollection.CountDocuments(ctx, queryDoc, &options.CountOptions{Collation: &options.Collation{Locale: "en", Strength: 2}})
	if err != nil {
		log.Print(err)
		return 0, nil, err
	}
	cursor, err := schematicCollection.Find(ctx, queryDoc, &options.FindOptions{Collation: &options.Collation{Locale: "en", Strength: 2}})
	if err != nil {
		log.Print(err)
		return 0, nil, err
	}
	defer cursor.Close(ctx)
	schematics := make([]model.Schematic, 0)
	count := 0
	docs := 0
	for cursor.Next(ctx) {
		if count >= offset {
			if docs < limit {
				var schematic model.Schematic
				if err = cursor.Decode(&schematic); err != nil {
					log.Print(err)
					return 0, nil, err
				} else {
					if !schematic.PrivateFile || schematic.Owner == owner {
						schematics = append(schematics, schematic)
						docs++
					}
				}
			} else {
				break
			}
		}
		count++
	}
	return n, schematics, nil
}

// GetSchematicsCount getting a sdingle schematic
func (m *MongoDAO) GetSchematicsCount(query string, owner string) (int64, error) {
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	schematicCollection := m.database.Collection(schematicsCollectionName)
	queryDoc := bson.M{}

	if query != "" {
		var queryM map[string]interface{}
		err := json.Unmarshal([]byte(query), &queryM)
		if err != nil {
			log.Print(err)
			return 0, err
		}
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
		data, _ := json.Marshal(queryDoc)
		log.Printf("mongoquery: %s\n", string(data))
	}
	n, err := schematicCollection.CountDocuments(ctx, queryDoc, &options.CountOptions{Collation: &options.Collation{Locale: "en", Strength: 2}})
	if err != nil {
		log.Print(err)
		return 0, err
	}
	return n, nil
}
*/

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

// GetUser getting the usermolde
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
			log.Print(err)
			return nil, err
		} else {
			user.Password = ""
			users = append(users, user)
		}
	}
	return users, nil
}

// GetUser getting the usermolde
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
	return nil, ErrNotImplemented
}

func (m *MongoDAO) Query(route model.Route, query string, offset int, limit int) (int, []model.JsonMap, error) {
	return 0, nil, ErrNotImplemented
}

func (m *MongoDAO) UpdateModel(route model.Route, data model.JsonMap) error {
	return ErrNotImplemented
}

func (m *MongoDAO) DeleteModel(route model.Route, dataId string) error {
	return ErrNotImplemented
}

// Ping pinging the mongoDao
func (m *MongoDAO) Ping() error {
	if !m.initialised {
		return errors.New("mongo client not initialised")
	}
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	return m.database.Client().Ping(ctx, nil)
}

// DropAll dropping all data from the database
func (m *MongoDAO) DropAll() {
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	collectionNames, err := m.database.ListCollectionNames(ctx, bson.D{}, &options.ListCollectionsOptions{})
	if err != nil {
		log.Fatal(err)
	}
	for _, name := range collectionNames {
		if name != usersCollectionName {
			collection := m.database.Collection(name)
			err = collection.Drop(ctx)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

// Stop stopping the mongodao
func (m *MongoDAO) Stop() {
	m.ticker.Stop()
	m.done <- true
}
