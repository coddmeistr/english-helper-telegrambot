package db

import (
	"context"
	"math/rand"

	"github.com/maxik12233/english-helper-telegrambot/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type IRepository interface {
	CreateMessage(msg *Message) error
	CreateTranslation(trnsl *Translation) error
	GetRandomTranslation() (*Translation, error)
	CreateConfig(cfg *Config) error
	GetConfig(userid uint) (*Config, error)
	UpdateConfig(cfg *Config) error
}

type MongoRepo struct {
	mongo *mongo.Database
}

func NewMongoRepo(mongo *mongo.Database) IRepository {
	return &MongoRepo{
		mongo: mongo,
	}
}

func (r *MongoRepo) CreateMessage(msg *Message) error {

	_, err := r.mongo.Collection("messages").InsertOne(context.TODO(), msg)
	if err != nil {
		return err
	}

	return nil
}

func (r *MongoRepo) CreateTranslation(trnsl *Translation) error {

	_, err := r.mongo.Collection("translations").InsertOne(context.TODO(), trnsl)
	if err != nil {
		return err
	}

	return nil
}

func (r *MongoRepo) GetRandomTranslation() (*Translation, error) {
	log := logger.GetLogger()

	res, err := r.mongo.Collection("translations").Find(context.TODO(), bson.D{})
	if err != nil {
		log.Error("Error while updating user config", zap.Error(err))
		return nil, err
	}

	var translations []Translation
	i := rand.Intn(res.RemainingBatchLength())
	if err = res.All(context.TODO(), &translations); err != nil {
		log.Error("Error while updating user config", zap.Error(err))
		return nil, err
	}

	return &translations[i], nil
}

func (r *MongoRepo) CreateConfig(cfg *Config) error {

	_, err := r.mongo.Collection("userconfigs").InsertOne(context.TODO(), cfg)
	if err != nil {
		return err
	}

	return nil
}

func (r *MongoRepo) GetConfig(userid uint) (*Config, error) {

	res := r.mongo.Collection("userconfigs").FindOne(context.TODO(), bson.D{{Key: "userid", Value: userid}})
	if res.Err() != nil {
		return nil, res.Err()
	}

	var cfg Config
	err := res.Decode(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (r *MongoRepo) UpdateConfig(cfg *Config) error {
	log := logger.GetLogger()

	update := bson.D{{Key: "$set", Value: cfg}}

	_, err := r.mongo.Collection("userconfigs").UpdateOne(context.TODO(), bson.D{{Key: "userid", Value: cfg.UserID}}, update)
	if err != nil {
		log.Error("Error while updating user config", zap.Error(err))
		return err
	}

	return nil
}
