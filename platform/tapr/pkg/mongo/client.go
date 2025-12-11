package mongo

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"bytetrade.io/web3os/tapr/pkg/apis/apr/v1alpha1"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"k8s.io/klog/v2"
)

type MongoClient struct {
	User     string
	Password string
	Database string
	Addr     string
	client   *mongo.Client
}

func (m *MongoClient) Connect(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	dsn := fmt.Sprintf("mongodb://%s:%s@%s/%s", m.User, m.Password, m.Addr, m.Database)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dsn))
	if err != nil {
		return err
	}

	m.client = client

	return client.Ping(ctx, nil)
}

func (m *MongoClient) Close(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}

func (m *MongoClient) CreateOrUpdateUserWithDatabase(ctx context.Context, user, pwd string, db []v1alpha1.MongoDatabase) error {
	if len(db) == 0 {
		return errors.New("db not specified")
	}

	adminDb := m.client.Database("admin")
	// auth user to every single db
	for _, authDB := range db {

		usersCollection := adminDb.Collection("system.users")
		query := bson.D{{Key: "user", Value: user}, {Key: "db", Value: authDB.Name}}
		var result bson.M
		err := usersCollection.FindOne(ctx, query).Decode(&result)

		getRoles := func() []bson.M {
			var res []bson.M
			for _, d := range db {
				res = append(res, bson.M{"role": "readWrite", "db": d.Name})
			}

			return res
		}

		var cmd bson.D
		if errors.Is(err, mongo.ErrNoDocuments) { // new user
			cmd = bson.D{
				{Key: "createUser", Value: user},
				{Key: "pwd", Value: pwd},
				{Key: "roles", Value: getRoles()},
			}
		} else if err != nil {
			return err
		} else {
			// update user
			cmd = bson.D{
				{Key: "updateUser", Value: user},
				{Key: "pwd", Value: pwd},
				{Key: "roles", Value: getRoles()},
			}
		}

		database := m.client.Database(authDB.Name) // create user for every db
		cmdResult := database.RunCommand(ctx, cmd)
		klog.Info("create or update mongodb user, ", cmdResult)
		if err = cmdResult.Err(); err != nil {
			return err
		}

		var res bson.M
		err = cmdResult.Decode(&res)
		if err != nil {
			klog.Error("decode mongo result error, ", res)
			return err
		}

		if res["ok"] == nil || res["ok"].(float64) != 1 {
			return errors.New(res["errmsg"].(string))
		}

		if len(authDB.Scripts) > 0 {
			err = m.ExecuteScript(authDB.Scripts, authDB.Name, user)
			if err != nil {
				klog.Error(err)
				return err
			}
		}

	} // end db loops

	return nil
}

func (m *MongoClient) DropUserAndDatabase(ctx context.Context, user string, db []v1alpha1.MongoDatabase) error {
	if user != "" {
		database := m.client.Database("admin")

		usersCollection := database.Collection("system.users")
		query := bson.D{{Key: "user", Value: user}}
		var result bson.M
		err := usersCollection.FindOne(ctx, query).Decode(&result)

		if errors.Is(err, mongo.ErrNoDocuments) { // new user
			klog.Info("user not found")
		} else if err != nil {
			return err
		} else {
			cmd := bson.D{
				{Key: "dropUser", Value: user},
			}

			if len(db) > 0 {
				for _, authDB := range db {
					database = m.client.Database(authDB.Name)
					cmdResult := database.RunCommand(ctx, cmd)
					klog.Info("drop mongodb user, ", cmdResult)
					if err = cmdResult.Err(); err != nil {
						return err
					}
				}
			} else {
				cmdResult := database.RunCommand(ctx, cmd)
				klog.Info("drop mongodb user, ", cmdResult)
				if err = cmdResult.Err(); err != nil {
					return err
				}
			}
		}
	}

	if len(db) > 0 {
		for _, d := range db {
			database := m.client.Database(d.Name)
			klog.Info("drop database, ", d.Name)
			err := database.Drop(ctx)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *MongoClient) ExecuteScript(scripts []string, databaseName, dbUsername string) error {
	dsn := fmt.Sprintf("mongodb://%s", m.Addr)
	var sb strings.Builder

	for _, cmd := range scripts {
		cmd = strings.ReplaceAll(cmd, "$databasename", databaseName)
		cmd = strings.ReplaceAll(cmd, "$dbusername", dbUsername)
		sb.WriteString(cmd)
		if !strings.HasSuffix(cmd, ";") {
			sb.WriteString(";")
		}
	}
	return m.eval(dsn, sb.String(), m.User, m.Password)
}

func (m *MongoClient) eval(dsn, scripts string, dbUsername, pwd string) error {
	cmd := exec.Command("mongosh", dsn, "--eval", fmt.Sprintf("\"%s\"", scripts), "-u", dbUsername, "-p", pwd)
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return errors.New(fmt.Sprintf("execute scripts failed with %v: %v", err, stderr.String()))
	}
	return nil
}
