# Fast-Quadric-Mesh-Simplification

this go implementaton is fast and versatil. port from [Sven FORSTMANN](https://github.com/sp4cerat/Fast-Quadric-Mesh-Simplification/tree/master). 
* memory efficient,
* free and for
* high quality

### go
5000 trias (692.6494ms)             |  10000 trias (769.2492ms)          | 20000 trias (747.3038ms)
:----------------------------------:|:----------------------------------:|:-------------------------:
![](snapshot/wall5000.png00.png)    |  ![](snapshot/wall10000.png01.png) | ![](snapshot/wall20000.png02.png) 

### cpp [link](https://github.com/sp4cerat/Fast-Quadric-Mesh-Simplification/) 
5000 trias (470.0ms)                |  10000 trias (445.0ms)             | 20000 trias (423.0ms)
:----------------------------------:|:----------------------------------:|:-------------------------:
![](snapshot/cpp500000.png)    |  ![](snapshot/cpp1000001.png) | ![](snapshot/cpp2000002.png) 


## Compiling and Usage
```shell
go build
./main.exe --input wall.obj --output wall20000.obj --target 2000
```

## Acknowledge
* https://github.com/sp4cerat/Fast-Quadric-Mesh-Simplification/tree/master
* github.com/udhos/gwob
* many good golang libs

