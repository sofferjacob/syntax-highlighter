# Actividad Integradora 3.4: Resaltador de Sintaxis

### Jacobo Soffer Levy | A01028653

## Instrucciones

Una vez que se carga el programa con `iex -S mix`, se puede ejecutar el resaltador de sintaxis de la siguiente forma:

```
iex> Hl.run(<input_file>, <output_file>)
```

Donde `input_file` es un archivo con código fuente en Go y `output_file` es el archivo al cual escribir el HTML resultante. Este archivo debe estar en el directorio root del proyecto para que tenga acceso al CSS que lo complementa. En esta repo se incluyen tres archivos que se pueden usar para probar el resaltador:

- `simple.go`
- `large.go`
- `larger.go`

## Expresiones regulares usadas

Las expresiones regulares usadas en este programa se pueden obtener ejecutando las siguientes funciones:
| **Tipo de Token** | **Función** |
|-------------------|------------------|
| Operador | `Hl.operators/0` |
| Keyword | `Hl.keywords/0` |
| Tipo | `Hl.types/0` |
| Variable | `Hl.vars/0` |
| Número | `Hl.numbers/0` |
| String | `Hl.strings/0` |

## Funcionamiento del programa

Una vez que se lee el archivo fuente, este se separa en cada espacio y new line (a excepción de las partes que están entre comillas) resultando en una lista de tokens. Esto se hace usando la función `Hl.split/1`, la cual tiene una complejidad temporal de O(n). Posteriormente en la función `Hl.run_regex/1` se aplican una serie de expresiones regulares a cada elemento de la lista, y en caso de que haya un match, se convierte en un elemento `<span>` con una clase CSS que identifica el tipo de token. Para el funcionamiento de esta función, se usan las funciones `Regex.match?/1` y `Regex.replace/3` en cada token para identificarlos, las cuales tienen una complejidad temporal de O(n), por lo que la complejidad temporal de la función `Hl.run_regex/1` es O(n^2) y por ende la complejidad temporal de la función `Hl.run/2` es O(n^2).

## Reflexión

Sin duda este fue un proyecto desafiante, sin embargo, me ayudó a aprender y desarrollar mis habilidades en Elixir y expresiones regulares. Elixir al ser un lenguaje funcional me obligó a cambiar mi forma de pensar al programar, y esto fue de las partes más desafiantes de este proyeto, junto con desarrollar expresiones regulares que pudieran identificar los diferentes tipos de tokens que existen en Go. El desarrollo
del algoritmo usado en el programa también fue desafiante, pero creo que despues de varias iteraciones llegué a una implementación simple y eficiente de este. Aunque no se usó un algoritmo con complejidad temporal de O(n), la implementación sigue siendo eficiente y capaz de procesar archvos grandes de forma rápida.
