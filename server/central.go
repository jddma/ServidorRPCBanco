package server

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strconv"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Central struct {
	db Database
	tokensKey []byte
}

//Método para generar el token para la sucursal
func (c * Central) generateKeyJWT()  {

	var key [32]byte
	_, err := io.ReadFull(rand.Reader, key[:])
	if err != nil {
		log.Fatal(err)
	}
	c.tokensKey = key[:]

}

//Método para registrar en el sistema una nueva sucursal
func (c *Central) registerNewSubsidiary()  {

	//Leer los datos del usuario
	var name string
	fmt.Print("Digite el nombre de la nueva sucursal: ")
	fmt.Scanln(&name)
	fmt.Print("Digite la dirección de la nueva sucursal: ")
	reader := bufio.NewReader(os.Stdin)
	address, _ := reader.ReadString('\n')
	address = strings.Replace(address, "\n","", -1)

	//Generar el token para la sucursal
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"name": name,
		"address": address,
	})
	tokenString, err := token.SignedString(c.tokensKey)

	//Datos a persistir
	newBranch := map[string]string {
		"name": name,
		"address": address,
		"token": tokenString,
	}

	//Persistir los datos en la base de datos
	collection := c.db.connection.Database("centralBank").Collection("branches")
	_, err = collection.InsertOne(context.TODO(), newBranch)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("TOKEN: ", tokenString)

}

//Método para opciones del usuario del servidor
func (c *Central) userOptions()  {

	active := true
	for active{
		fmt.Print("***Opciones***\n\t* Registrar nueva sucursal [1]\nDigite un opción: ")

		var option int
		fmt.Scanf("%d", &option)

		switch option {

		case 1:
			c.registerNewSubsidiary()
			break

		}
	}

}

//Método para decodificar la información de las operaciones en fomato JSON
func (c *Central) decodeOperationData(operationData string) map[string]string {

	var result map[string]string

	err := json.Unmarshal([]byte(operationData), &result)
	if err != nil {
		log.Fatal(err)
	}

	return result

}

//Método remoto para validar una sucursal que empieza operaciones
func (c *Central) ValidateBranch(token string, response *bool) error {

	//Establecer la condición de filtro
	filter := bson.D{{
		"token", token,
	}}

	//Realizar la query
	collection := c.db.connection.Database("centralBank").Collection("branches")
	var queryResult Branch
	err := collection.FindOne(context.TODO(), filter).Decode(&queryResult)
	if err != nil {
		*response = false
		return nil
	}

	*response = true
	return nil

}

//Método para retirar dinero
func (c* Central) Withdrawals(operationDataArgs string, response *bool) error {


	var operationData string = string(operationDataArgs)
	//Obtener la información de la operación
	operationDataMap := c.decodeOperationData(operationData)

	//Enstablecer el cirterio de filtro
	filter := bson.D{{
		"document", operationDataMap["document"],
	}}

	//Realizar la query para obtener el monto actual
	collection := c.db.connection.Database("centralBank").Collection("clients")
	var queryResult client
	err := collection.FindOne(context.TODO(), filter).Decode(&queryResult)
	if err != nil {
		*response = false
		return nil
	}
	//Establecer la respueta
	currentBalance := queryResult.Mount

	mountToAdd, _ := strconv.Atoi(operationDataMap["mountToRemove"])
	newBalance := currentBalance - mountToAdd

	//Establecer los nuevos datos
	update := bson.D{primitive.E{Key: "$set", Value: bson.D{
		primitive.E{Key: "mount", Value: newBalance},
	}}}

	t := &client{}

	err = collection.FindOneAndUpdate(context.TODO(), filter, update).Decode(t)
	if err != nil {
		log.Println(err)
		*response = false
		return nil
	}

	*response = true
	return nil

}

//Método remoto para realziar una consignación
func (c* Central) AddMoney(operationData string, response *bool) error {

	//Obtener la información de la operación
	operationDataMap := c.decodeOperationData(operationData)

	//Enstablecer el cirterio de filtro
	filter := bson.D{{
		"document", operationDataMap["document"],
	}}

	//Realizar la query para obtener el monto actual
	collection := c.db.connection.Database("centralBank").Collection("clients")
	var queryResult client
	err := collection.FindOne(context.TODO(), filter).Decode(&queryResult)
	if err != nil {
		*response = false
		return nil
	}
	//Establecer la respueta
	currentBalance := queryResult.Mount

	mountToAdd, _ := strconv.Atoi(operationDataMap["mountToAdd"])
	newBalance := currentBalance + mountToAdd

	//Establecer los nuevos datos
	update := bson.D{primitive.E{Key: "$set", Value: bson.D{
		primitive.E{Key: "mount", Value: newBalance},
	}}}

	t := &client{}

	err = collection.FindOneAndUpdate(context.TODO(), filter, update).Decode(t)
	if err != nil {
		log.Println(err)
		*response = false
		return nil
	}

	*response = true
	return nil

}

//Método remoto para modificar una cuenta
func (c *Central) ModifyAccount(operationData string, response *bool) error{

	//Obtener la información de la operación
	operationDataMap := c.decodeOperationData(operationData)

	//Establecer el cirterio de filtro
	filter := bson.D{{
		"document", operationDataMap["document"],
	}}

	//Establecer los nuevos datos
	update := bson.D{primitive.E{Key: "$set", Value: bson.D{
		primitive.E{Key: "document", Value: operationDataMap["newDocument"]},
	}}}

	t := &client{}

	//Realizar la modificación a la base de datos
	collection := c.db.connection.Database("centralBank").Collection("clients")
	err := collection.FindOneAndUpdate(context.TODO(), filter, update).Decode(t)
	if err != nil {
		log.Println(err)
		*response = false
		return nil
	}

	*response = true
	return nil

}

//Método remoto para eliminar una cuenta
func (c *Central) DeleteAccount(operationData string, response *bool) error {

	operationDataMap := c.decodeOperationData(operationData)

	filter := bson.D{{
		"document", operationDataMap["document"],
	}}

	//Realizar la query
	collection := c.db.connection.Database("centralBank").Collection("clients")

	_, err := collection.DeleteOne(context.TODO(), filter)

	//Establecer la respueta
	if err != nil{
		log.Println(err)
		*response = false
		return nil
	}

	*response = true
	return nil
}

//Método remoto para obtener el monto de la cuenta de un usuario
func (c *Central) GetBalance(operationData string, response *int) error {

	//Decodificar el argumento de información de la operación en formato JSON
	operationDataMap := c.decodeOperationData(operationData)

	//Establecer la condición de filtro
	filter := bson.D{{
		"document", operationDataMap["document"],
	}}

	//Realizar la query
	collection := c.db.connection.Database("centralBank").Collection("clients")
	var queryResult client
	err := collection.FindOne(context.TODO(), filter).Decode(&queryResult)
	if err != nil {
		*response = 0
		return nil
	}
	//Establecer la respueta
	*response = queryResult.Mount

	return nil

}

//Método para inciar el servidor
func (c *Central) StartServer(port string)  {

	//Instanciar el servidor de rpc
	rpc.Register(c)
	rpc.HandleHTTP()

	l, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal("Listen error: ", err)
	}

	c.db.openConnection()
	c.generateKeyJWT()

	//Iniciar una rutina para ejecutar el servidor y las opciones del mismo
	go c.userOptions()

	//Encender el servidor
	fmt.Println("Servidor escuchando el puerto ", port, "\n")
	http.Serve(l, nil)

}
