package infisical

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"k8s.io/klog/v2"
)

type MongoClient struct {
	User     string
	Password string
	Database string
	Addr     string
}

type UserMDB struct {
	ID                  primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty" mapstructure:"_id"`
	Email               string             `bson:"email" json:"email" mapstructure:"email"`
	FirstName           string             `bson:"firstName" json:"firstName" mapstructure:"firstName"`
	LastName            string             `bson:"lastName" json:"lastName" mapstructure:"lastName"`
	EncryptionVersion   int32              `bson:"encryptionVersion" json:"encryptionVersion" mapstructure:"encryptionVersion"`
	ProtectedKey        string             `bson:"protectedKey" json:"protectedKey" mapstructure:"protectedKey"`
	ProtectedKeyIV      string             `bson:"protectedKeyIV" json:"protectedKeyIV" mapstructure:"protectedKeyIV"`
	ProtectedKeyTag     string             `bson:"protectedKeyTag" json:"protectedKeyTag" mapstructure:"protectedKeyTag"`
	PublicKey           string             `bson:"publicKey" json:"publicKey" mapstructure:"publicKey"`
	EncryptedPrivateKey string             `bson:"encryptedPrivateKey" json:"encryptedPrivateKey" mapstructure:"encryptedPrivateKey"`
	IV                  string             `bson:"iv" json:"iv" mapstructure:"iv"`
	Tag                 string             `bson:"tag" json:"tag" mapstructure:"tag"`
	Salt                string             `bson:"salt" json:"salt" mapstructure:"salt"`
	Verifier            string             `bson:"verifier" json:"verifier" mapstructure:"verifier"`
	RefreshVersion      int32              `bson:"refreshVersion" json:"refreshVersion" mapstructure:"refreshVersion"`
	IsMfaEnabled        bool               `bson:"isMfaEnabled" json:"isMfaEnabled" mapstructure:"isMfaEnabled"`
	MfaMethods          []string           `bson:"mfaMethods" json:"mfaMethods" mapstructure:"mfaMethods"`
}

type OrganizationMDB struct {
	ID   primitive.ObjectID `bson:"_id,omitempty" mapstructure:"_id"`
	Name string             `bson:"name"`
}

type MembershipOrgMDB struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" mapstructure:"_id"`
	Organization primitive.ObjectID `bson:"organization"`
	Role         string             `bson:"role"`
	Status       string             `bson:"status"`
	User         primitive.ObjectID `bson:"user"`
}

func (m *MongoClient) TryConnect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dsn := fmt.Sprintf("mongodb://%s:%s@%s/%s", m.User, m.Password, m.Addr, m.Database)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dsn))
	if err != nil {
		return err
	}
	defer client.Disconnect(ctx)

	return client.Ping(ctx, nil)
}

func (m *MongoClient) GetUser(basectx context.Context, email string) (*UserMDB, error) {
	if email == "" {
		return nil, errors.New("email is empty")
	}

	ctx, cancel := context.WithTimeout(basectx, 10*time.Second)
	defer cancel()

	dsn := fmt.Sprintf("mongodb://%s:%s@%s/%s", m.User, m.Password, m.Addr, m.Database)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dsn))
	if err != nil {
		return nil, err
	}
	defer client.Disconnect(ctx)

	db := client.Database(m.Database)

	var result bson.M
	singleRes := db.Collection("users").FindOne(ctx, bson.D{
		bson.E{Key: "email", Value: email},
	})
	if err := singleRes.Decode(&result); err != nil {
		return nil, err
	}

	var user UserMDB
	err = mapstructure.Decode(result, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (m *MongoClient) SaveUser(basectx context.Context, user *UserMDB) (string, error) {
	if user == nil {
		return "", errors.New("user is empty")
	}

	ctx, cancel := context.WithTimeout(basectx, 10*time.Second)
	defer cancel()

	dsn := fmt.Sprintf("mongodb://%s:%s@%s/%s", m.User, m.Password, m.Addr, m.Database)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dsn))
	if err != nil {
		return "", err
	}
	defer client.Disconnect(ctx)

	db := client.Database(m.Database)

	userresult, err := db.Collection("users").InsertOne(ctx, user)
	if err != nil {
		klog.Info("save user error, ", err)
		return "", err
	}

	dbOrg := &OrganizationMDB{
		Name: "Terminus",
	}
	orgRes, err := db.Collection("organizations").InsertOne(ctx, dbOrg)
	if err != nil {
		klog.Info("save org error, ", err)
		return "", err
	}

	dbMember := &MembershipOrgMDB{
		Organization: orgRes.InsertedID.(primitive.ObjectID),
		User:         userresult.InsertedID.(primitive.ObjectID),
		Role:         "owner",
		Status:       "accepted",
	}
	_, err = db.Collection("membershiporgs").InsertOne(ctx, dbMember)
	if err != nil {
		klog.Info("save membershiporgs error, ", err)
		return "", err
	}

	return userresult.InsertedID.(primitive.ObjectID).Hex(), nil
}
