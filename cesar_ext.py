import sys


# Símbolos que pueden cifrarse
ALFABETO = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz'

# Almacena la cadena cifrada/descifrada
salida = ''

# Guarda la opción deseada
modo = sys.argv[1]

# Se almacena el texto y la clave
clave = int(sys.argv[2])
texto = sys.argv[3]


# Ejecuta el proceso letra a letra
for simbolo in texto:
    if simbolo in ALFABETO:
        # Identifica la posición de cada símbolo
        pos = ALFABETO.find(simbolo)
        # y ejecuta la operación de cifrado/descifrado
        if modo == 'c':
            pos = (pos + clave) % len(ALFABETO)
        elif modo == 'D':
            pos = (pos - clave) % len(ALFABETO)

        # Añade el nuevo símbolo a la cadena
        salida += ALFABETO[pos]
        
    # Añade a la cadena el símbolo sin cifrar ni descifrar
    else:
        salida += simbolo

# Imprime en pantalla el resultado
print(salida)


