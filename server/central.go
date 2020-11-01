package server

import (
	"bufio"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strings"


	"github.com/dgrijalva/jwt-go"
)

type Central struct {
	db Database
	tokensKey []byte
}

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

		case 2:
			active = false
			break

		}
	}

}

func (c *Central) GetBalance(operationData string, response *bool) error {

	return nil

}

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
