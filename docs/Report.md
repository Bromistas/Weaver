# Informe Técnico del Proyecto de scrapper Web Distribuido

## Tabla de Contenidos

1. Introducción
2. Arquitectura
3. Comunicación
4. Coordinación
5. Descubrimiento de la Red
6. Replicación
7. Persistencia
8. Tolerancia a Fallos
9. Conclusión

---

## 1. Introducción

El objetivo de este proyecto es desarrollar un scrapper web distribuido que raspe eficientemente sitios web de comercio
electrónico en línea para encontrar los mejores productos que coincidan con una consulta dada. El sistema admite dos
operaciones principales: `scrap <query>` y `gather`. El comando `scrap` consulta varios sitios web de comercio
electrónico y raspa los mejores productos que coincidan con la consulta. El comando `gather` devuelve una tabla de todos
los productos raspados hasta ahora. La arquitectura comprende tres tipos de nodos: colas, scrappers y almacenes. Este
informe detalla los aspectos técnicos del proyecto, cubriendo la arquitectura, comunicación, coordinación,
descubrimiento de red, replicación, persistencia y tolerancia a fallos.

---

## 2. Arquitectura

La arquitectura del proyecto de scrapper web distribuido es una arquitectura de microservicios que consta de tres tipos
principales de nodos: colas, scrappers y almacenes.

### Colas

Las colas son responsables de almacenar mensajes que contienen URLs para raspar y el tipo de URL que son. Estas colas no
son implementaciones estándar como RabbitMQ; en su lugar, están construidas a medida para satisfacer las necesidades
específicas de este proyecto. Cada cola permite poner y sacar mensajes con un sistema de bloqueo distribuido para
asegurar que no dos scrappers lean el mismo mensaje simultáneamente. Cuando se saca un mensaje, se vuelve invisible
durante hasta 40 segundos, asegurando que no pueda ser accedido por otro scrapper durante este período. Este mecanismo
previene la pérdida de URLs si un nodo scrapper falla antes de completar su tarea. Una solicitud de ACK marca un mensaje
como procesado exitosamente, eliminándolo de la cola.

### scrappers

Los scrappers actúan como nodos de transición en el sistema. Continuamente leen de todas las colas y realizan las
tareas de raspado. Un scrapper puede poner nuevas URLs en la cola si es necesario raspar más páginas o enviar los datos
de los productos raspados a los almacenes. Cada nodo scrapper contiene referencias a todas las colas y un nodo de
almacenamiento, permitiendo la lectura simultánea de múltiples colas en diferentes hilos. Este diseño permite la
escalabilidad, ya que se pueden añadir nodos scrappers adicionales para manejar la carga aumentada sin modificar la
estructura de las colas.

### Almacenes

Los almacenes están agrupados en un anillo CHORD, un sistema distribuido basado en hashing consistente. Cada nodo de
almacenamiento es responsable de almacenar una parte de los datos raspados. Cuando los datos se envían a los almacenes,
primero se dirigen a un nodo de punto de entrada, que luego los redirige al nodo apropiado en el anillo CHORD. Este
enfoque asegura que los datos se distribuyan equitativamente entre los nodos, previniendo que un solo nodo se convierta
en un cuello de botella. Además, el anillo CHORD proporciona una búsqueda y recuperación de datos eficientes, mejorando
el rendimiento y la fiabilidad del sistema.

---

## 3. Comunicación

La comunicación dentro del sistema de scrapper web distribuido se facilita a través de dos protocolos principales: gRPC
y HTTP. Estos protocolos sirven para diferentes propósitos y se eligen según sus fortalezas y adecuación para tareas
específicas.

### gRPC

gRPC se utiliza para manejar tareas relacionadas con CHORD, como encontrar los nodos sucesores y predecesores,
estabilizar la red y mantener el anillo de hashing consistente. Al utilizar gRPC sobre HTTP/2, el sistema se beneficia
de una comunicación de alto rendimiento y baja latencia con una eficiente serialización y deserialización de mensajes.
Este protocolo asegura que los nodos distribuidos en el anillo CHORD puedan comunicarse rápida y confiablemente entre
sí, manteniendo la integridad y estabilidad de la red.

### HTTP

HTTP se emplea para tareas específicas de los nodos, incluyendo operaciones de colas (como sacar y poner mensajes) y
replicación de datos en los almacenes. Usando HTTP/1.1, estas tareas se manejan de una manera sencilla y ampliamente
soportada, permitiendo una fácil integración con tecnologías web. La separación de tareas relacionadas con CHORD y
tareas específicas de los nodos entre gRPC y HTTP asegura que cada protocolo se utilice según sus fortalezas,
optimizando la eficiencia general de la comunicación del sistema.

### Multiplexado

El sistema utiliza un multiplexor para redirigir las solicitudes entrantes al servidor apropiado según el protocolo
utilizado. Tanto los servidores gRPC como HTTP funcionan en el mismo puerto, con el multiplexor dirigiendo las
solicitudes HTTP/1.1 al servidor HTTP y las solicitudes HTTP/2 al servidor gRPC. Este diseño simplifica la configuración
de la red y permite una gestión fluida de la comunicación, asegurando que el protocolo correcto maneje cada tipo de
solicitud.

---

## 4. Coordinación

La coordinación dentro del sistema de scrapper web distribuido es crucial para asegurar que los nodos operen
eficientemente y sin conflictos. El sistema emplea un mecanismo de bloqueo distribuido y un diseño escalable para
gestionar la coordinación entre nodos.

### Sistema de Bloqueo Distribuido

El sistema de bloqueo distribuido se implementa para controlar el acceso a las colas, asegurando que no dos nodos
scrappers intenten leer el mismo mensaje simultáneamente. Cuando se saca un mensaje de una cola, no se elimina
inmediatamente; en su lugar, se vuelve invisible durante hasta 40 segundos. Durante este período, el mensaje no puede
ser accedido por ningún otro scrapper, previniendo el procesamiento duplicado. Si un nodo scrapper procesa exitosamente
un mensaje, envía una solicitud de ACK a la cola, marcando el mensaje como procesado y eliminándolo de la cola. Este
enfoque mitiga el riesgo de perder URLs si un nodo scrapper falla antes de completar su tarea.

### Escalabilidad

La escalabilidad del sistema se logra a través del diseño de los scrappers y la estructura del anillo CHORD de los
almacenes. Los scrappers leen de todas las colas simultáneamente en diferentes hilos, permitiendo el procesamiento
eficiente de una gran cantidad de URLs. El sistema puede escalar fácilmente añadiendo más nodos scrappers, que
comenzarán a leer de las colas existentes sin requerir cambios en la estructura de las colas.

En el anillo CHORD, cada nodo de almacenamiento es responsable de una parte de los datos, y el mecanismo de hashing
consistente asegura que los datos se distribuyan equitativamente. La estructura del anillo permite una búsqueda y
recuperación de datos eficientes, incluso cuando aumenta el número de nodos. El uso de un único punto de entrada para
redirigir los datos al nodo de almacenamiento apropiado simplifica la coordinación y gestión del almacenamiento de
datos, asegurando que el sistema siga siendo eficiente y escalable.

---

## 5. Descubrimiento de la Red

El descubrimiento de la red es un aspecto esencial del sistema de scrapper web distribuido, permitiendo que los nodos se
encuentren y comuniquen entre sí de manera efectiva. El sistema utiliza un mecanismo de difusión para facilitar el
descubrimiento de la red y asegurar que los nodos puedan unirse y participar en la red rápidamente.

### Mecanismo de Difusión

Todos los nodos de almacenamiento y colas escuchan en una IP de difusión y en sus respectivas IPs de nodo. Cuando un
nuevo nodo desea unirse a la red o descubrir nodos existentes, envía una pregunta de difusión a la red. Los nodos que
reciben esta pregunta responden con sus respectivos roles (ya sea almacén o cola), permitiendo que el nuevo nodo los
identifique y se comunique con ellos.

### Proceso de Descubrimiento

El proceso de descubrimiento de la red asegura que los nodos puedan unirse y salir dinámicamente de la red sin
interrumpir la operación general. Cuando un nuevo nodo se une, utiliza el mecanismo de difusión para anunciar su
presencia y descubrir otros nodos. De manera similar, los nodos existentes pueden usar el mecanismo de difusión para
descubrir nuevos nodos y actualizar sus referencias internas. Este enfoque permite que la red se adapte a los cambios y
mantenga una comunicación eficiente entre los nodos.

---

## 6. Replicación

La replicación es un aspecto crítico del sistema de scrapper web distribuido, asegurando la persistencia y
disponibilidad de los datos incluso en caso de fallos de nodos. El sistema emplea una estrategia de replicación para los
nodos de almacenamiento, replicando datos para asegurar redundancia y tolerancia a fallos.

### Replicación de Datos

La replicación de datos es un requisito técnico para el sistema, asegurando que los datos raspados sean persistentes y
estén disponibles incluso si un nodo de almacenamiento falla. Cada dato se replica en un nodo adicional, específicamente
en el predecesor del nodo de almacenamiento responsable de los datos. Este enfoque asegura que al menos otro nodo tenga
una copia de los datos, proporcionando redundancia y tolerancia a fallos.

### Disparadores de Replicación

La replicación ocurre en dos escenarios principales:

1. **Replicación Inicial:** Cuando los datos se almacenan por primera vez en un nodo y aún no se han replicado, se
   replican inmediatamente en el nodo predecesor.
2. **Cambio de Predecesor:** Cuando cambia el predecesor de un nodo (debido a la entrada o salida de un nodo en la red),
   los datos se replican al nuevo predecesor para asegurar consistencia y disponibilidad.

Esta estrategia de replicación garantiza que los datos siempre se almacenen en al menos dos nodos, mejorando la
fiabilidad y la tolerancia a fallos del sistema.

---

## 7. Persistencia

La persistencia es esencial para asegurar que los datos raspados por el sistema se almacenen de manera fiable y
permanezcan accesibles a lo largo del tiempo. El sistema emplea diferentes estrategias de persistencia para las colas y
los almacenes, reflejando sus respectivos roles y requisitos.

### Persistencia de Colas

Las colas no proporcionan replicación o persistencia de datos, ya que los mensajes en las colas son efímeros por diseño.
Si una cola muere, los mensajes que contiene se pierden. Sin embargo, dado que todos los scrappers leen de todas las
colas, el impacto de perder un mensaje se minimiza. Este enfoque simplifica el diseño de las colas y evita la
complejidad de manejar mensajes duplicados, lo que requeriría hacer que los scrappers fueran con estado y aumentaría
significativamente la sobrecarga.

### Persistencia de Almacenes

En contraste, los nodos de almacenamiento proporcionan una persistencia de datos robusta. Los datos almacenados en el
anillo CHORD se replican para asegurar que permanezcan disponibles incluso si un nodo falla. Esta estrategia de
persistencia asegura que los datos raspados no se pierdan y puedan ser recuperados de manera fiable cuando sea
necesario. El uso de hashing consistente y replicación de datos garantiza que los datos se distribuyan equitativamente y
se almacenen de manera redundante en toda la red.

---

## 8. Tolerancia a Fallos

La tolerancia a fallos es un aspecto crítico del sistema de scrapper web distribuido, asegurando que el sistema pueda
seguir operando incluso en presencia de fallos de nodos. El sistema emplea diferentes mecanismos de tolerancia a fallos
para las colas y los almacenes, reflejando sus respectivos roles y requisitos.

### Tolerancia a Fallos de Colas

La tolerancia a fallos de las colas se logra a través del sistema de bloqueo distribuido y la naturaleza efímera de los
mensajes. Cuando se saca un mensaje de una cola, se vuelve invisible durante hasta 40 segundos, previniendo que otros
scrappers lo accedan durante este período. Si un nodo scrapper falla antes de completar su tarea, el mensaje
eventualmente se volverá visible nuevamente y podrá ser procesado por otro scrapper. Este enfoque asegura que los
mensajes no se pierdan debido a fallos de nodos, aunque los mensajes en sí mismos no se replican ni persisten.

### Tolerancia a Fallos de Almacenes

La tolerancia a fallos de los almacenes se logra a través de la replicación de datos. Cada dato se replica en el
predecesor del nodo de almacenamiento responsable de los datos, asegurando que al menos otro nodo tenga una copia. Esta
estrategia de replicación asegura que los datos permanezcan disponibles incluso si un nodo de almacenamiento falla. En
caso de fallo de un nodo, los datos replicados aseguran que el sistema pueda seguir operando sin perder ninguna
información.

### Manejo de Fallos de Nodos

El sistema está diseñado para manejar fallos de nodos de manera eficiente. Para las colas, la naturaleza efímera de los
mensajes y el sistema de bloqueo distribuido minimizan el impacto de perder un mensaje. Para los almacenes, la
estrategia de replicación asegura que los datos permanezcan disponibles incluso si un nodo falla. Al replicar datos en
el nodo predecesor, el sistema asegura que siempre haya una copia redundante de los datos, mejorando su tolerancia a
fallos y fiabilidad.

---

## 9. Conclusión

El proyecto de scrapper web distribuido proporciona una solución robusta y escalable para raspar sitios web de comercio
electrónico y recopilar datos de productos. La arquitectura, los protocolos de comunicación, los mecanismos de
coordinación, el descubrimiento de red, la replicación, la persistencia y las estrategias de tolerancia a fallos
aseguran una operación eficiente y confiable. Al aprovechar Go por su manejo ligero y eficiente de procesos
distribuidos, hilos y mecanismos de bloqueo, el proyecto logra sus objetivos de manera efectiva. Este informe sirve como
una guía comprensiva de los aspectos técnicos del proyecto, proporcionando una comprensión clara de su diseño e
implementación.

La combinación de colas construidas a medida, nodos scrappers escalables y un almacén basado en un anillo CHORD asegura
que el sistema pueda manejar tareas de raspado a gran escala de manera eficiente. El uso de gRPC y HTTP para la
comunicación, junto con una estrategia robusta de replicación y tolerancia a fallos, asegura que el sistema siga siendo
confiable y funcione bien incluso ante fallos de nodos. Este proyecto de scrapper web distribuido demuestra la
efectividad de un sistema distribuido bien diseñado para resolver tareas complejas de raspado web.