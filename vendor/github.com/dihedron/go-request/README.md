# go-request
This project implements a simple HTTP requests builder with a fluent API; rrequest submission and response handling is out of scope.

## Usage
The library can be imported via
``` golang {.line-numbers}
import (
	"github.com/dihedron/go-request"
)
```
and provides a simple factory pattern that can be used to add information and finally retrieve an ```http.Request``` object. The ```Builder``` is available under the ```request``` namespace.
Typical usage would be along the following lines:
``` golang {.line-numbers}
req, _ := request.
	New("").                                        // create the factory
	Post().                                         // POST HTTP method
	UserAgent("myUserAgent/1.0").                   // sets the user agent
	Base("https://www.example.com/").               // base URL for requests
	Path("api/v2/login?param1=value1").             // extra request path
	Add().                                          // start adding
	QueryParameter("param1", "value1a", "value1b").	// adds a query param to URL
	QueryParameter("param2", "value2").             // adds another query param
	Header("X-Auth-Token", "1234567890abcdef").     // adds a header
	WithJSONEntity(myTaggedStruct).                 // adds the request body from a struct
	Make()
```
Let's break it apart, line by line: 
1. the first instruction (```New("")```) creates a new builder; this method can be used to pass in the request's URL, or it can be specified later, as in this example;
2. the HTTP method is ```POST```; if not specified, the default is ```GET```;
3. the second instruction (```UserAgent()```) adds the ```User-Agent``` header to the request; a special facility is provided for the ```User-Agent``` and ```Content-Type``` headers since these are more common used than others;
4. the ```Base()``` call sets the base URL for requests generated from this builder; this can be very useful when creating sub-builders, because they will all share the same base URL and have different paths;
5. ```Path()``` sets the resource path; paths can be absolute (in which case the base path should have a trailing slash) or relative and include ```../```; if the path includes query parameters, they will be preserved when the request is generated;
6. ```Add()``` opens a section where the builder accepts query parameter and header values that will be __added__ to the request; other accepted operations are ```Set()``` (which __replaces__ query parameters and headers if already present), ```'Del()``` (which __removes__ headers and query parameter with the given key), ```Remove()``` (which __removes__ headers and query parameters whose keys match the given regular expression);
7. ```QueryParameter()``` and ```Header()``` are used to specify query parameters and headers, respectively, that will be added to the builder;
8. ```WithJSONEntity()``` (and its XML counterpart ```WithXMLEntity()```) is a way to add the request entity (payload) by passing in a tagged struct; all fields marked with ```json``` (and ```xml```) will be stored as part of the JSON (XML) request body; these methods also have the side effect of setting the ```USer-Agent``` if none was set already;
9. ```Make()``` creates the ```http.Request```.
 
The library provides the following additional facilities:
- reading of raw data into the request body (see ```WithEntity(io.Reader)```), as follows:
``` golang {.line-numbers}
file, _ := os.Open("path/file.ext")
req, _ := request.
	New("").
	// more methods here...
	WithEntity(bufio.NewReader(file))
```
- populating headers ad query parameters from a struct, whose fields are tagged with ```header``` and ```parameter``` tags respectively, or from a ```map[string][]string``` (see ```HeaderFrom()``` and ```QueryParametersFrom()```).
``` golang {.line-numbers}
req, _ := request.
	New("").
	// more methods here...
	Set().
	QueryParametersFrom(myMapQP).
	QueryParametersFrom(myStructQP).
	HeadersFrom(&myMapH).
	HeadersFrom(&myStructH).
	Make()
```
Note from the example that both ```struct```, ```map[string][]string``` and their pointers are supported.

A ```Builder``` can be used to create sub-```Builder```s that has a copy of the parent's headers and query parameters at that moment, plus a shared reference to the entity ```ioReader```:
``` golang {.line-numbers}
parent, _ := request.
	New("").
	Base("https://www.example.com/")
	// more methods here...
child := parent.New() // shares headers and query parameters.
```	


## Contributing
All contributions are welcome provided they don't spoil the simplicity of the API and that complete coverage with automatic __unit tests__ is provided.