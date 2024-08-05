package controller

import (
	"context"
	"fmt"
	"golang-restaurant-management/database"
	"golang-restaurant-management/models"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

var orderCollection *mongo.Collection = database.OpenCollection(database.Client, "order")

func GetOrders() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		result, err := orderCollection.Find(context.TODO(), bson.M{})
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing order items"})
		}
		var allOrders []bson.M
		if err = result.All(ctx, &allOrders); err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, allOrders)
	}
}

func GetOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		orderId := c.Param("order_id")
		var order models.Order

		err := orderCollection.FindOne(ctx, bson.M{"order_id": orderId}).Decode(&order)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while fetching the orders"})
		}
		c.JSON(http.StatusOK, order)
	}
}

// func CreateOrder() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		var table models.Table
// 		var order models.Order

// 		if err := c.BindJSON(&order); err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 			return
// 		}

// 		validationErr := validate.Struct(order)

// 		if validationErr != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
// 			return
// 		}

// 		if order.Table_id != nil {
// 			err := tableCollection.FindOne(ctx, bson.M{"table_id": order.Table_id}).Decode(&table)
// 			defer cancel()
// 			if err != nil {
// 				msg := fmt.Sprintf("message:Table was not found")
// 				c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
// 				return
// 			}
// 		}

// 		order.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
// 		order.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

// 		order.ID = primitive.NewObjectID()
// 		order.Order_id = order.ID.Hex()

// 		result, insertErr := orderCollection.InsertOne(ctx, order)

// 		if insertErr != nil {
// 			msg := fmt.Sprintf("order item was not created")
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
// 			return
// 		}

//			defer cancel()
//			c.JSON(http.StatusOK, result)
//		}
//	}
func CreateOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var table models.Table
		var order models.Order

		if err := c.BindJSON(&order); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate order
		validationErr := validate.Struct(order)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		// Check if table exists
		if order.Table_id != nil {
			err := tableCollection.FindOne(ctx, bson.M{"table_id": *order.Table_id}).Decode(&table)
			if err != nil {
				// Log and return specific error message
				log.Printf("Table lookup failed for ID %s: %v", *order.Table_id, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Table was not found"})
				return
			}
		}

		// Create order
		order.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		order.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		order.ID = primitive.NewObjectID()
		order.Order_id = order.ID.Hex()

		result, insertErr := orderCollection.InsertOne(ctx, order)
		if insertErr != nil {
			log.Printf("Order creation failed: %v", insertErr)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Order item was not created"})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func UpdateOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		var table models.Table
		var order models.Order

		var updateObj primitive.D

		orderId := c.Param("order_id")
		if err := c.BindJSON(&order); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Check if table exists and use correct field name
		if order.Table_id != nil {
			err := tableCollection.FindOne(ctx, bson.M{"table_id": *order.Table_id}).Decode(&table)
			defer cancel()
			if err != nil {
				msg := fmt.Sprintf("Table was not found for ID %s", *order.Table_id)
				c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
				return
			}
			updateObj = append(updateObj, bson.E{"table_id", *order.Table_id})
		}

		// Update the `Updated_at` field
		order.Updated_at = time.Now()
		updateObj = append(updateObj, bson.E{"updated_at", order.Updated_at})

		// Perform the update operation
		upsert := true
		filter := bson.M{"order_id": orderId}
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		result, err := orderCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{"$set", updateObj}, // Fixed typo here
			},
			&opt,
		)

		if err != nil {
			msg := fmt.Sprintf("Order update failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func OrderItemOrderCreator(order models.Order) string {

	order.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	order.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	order.ID = primitive.NewObjectID()
	order.Order_id = order.ID.Hex()

	orderCollection.InsertOne(ctx, order)
	defer cancel()

	return order.Order_id
}
