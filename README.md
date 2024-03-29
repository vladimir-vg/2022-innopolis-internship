# Idea

Software projects tend to have thousands of lines of code which are often hard to navigate. Since we often think about programs as black boxes, it is easier to start learning system from parts that handle input and output (IO). It might be useful to have a tool that provides bird's view of the code base, highlighting locations that are related to IO.

Contemporary software systems are often written with concurrent computation in mind, in order to utilize modern multicore CPUs. As a result, structure of software system is often involves several processes or threads organized into a tree. The structure of such composition is often not immediately clear from the source code. It might be useful to display structure of such composition together with IO-related source code locations.

In order to create such tool we need to learn which source files are going to be included in the final executable. Since many languages use various building systems with different source code organization conventions, use of such tool may require ad-hoc configuration for each project.

Another problem is that tree of processes/threads is formed dynamically and often impossible to derive by mere static analysis. It seems that creating such tool might be a great challenge.

However, if we focus only on software written in Go language then it may be possible. Go language has strict standartized build pipeline, which removes any need for ad-hoc configuration for each project. Another advantage is that concurrency in Go is well structured via goroutines and channels that have distinct syntax, which greatly simplifies analysis.

The aim of this project is to provide a web-based visual UI for navigation in software projects written in Go language.

# User Interface

We need to navigate around goroutines, channels and IO calls. In that case it would be reasonable to represent collected information on a 2D plane, with time on Y axis and classes of goroutines on X axis, similar to classical Sequence Diagrams: https://en.wikipedia.org/wiki/Sequence_diagram

Unlike actors in Sequence Diagrams, goroutines have hierarchical nature: goroutines may spawn other goroutines. Thus, whole system have to have tree-like structure, which make take a lot of space to represent properly. It might be reasonable to fold parts of the tree, similar to what is done is this demo: https://www.youtube.com/watch?v=7KMezzzsRY8

![sketch of user interface](ui-sketch.png "sketch of user interface")

User will see the tree of goroutines, a directed spawn graph. Each goroutine (node) will have marks signifying IO calls or channel send/receive statements that can be displayed on the right on selection.

# Main parts

Go language already has well developed libraries for static code analysis that allow to traverse the code: https://pkg.go.dev/golang.org/x/tools@v0.1.10/go/analysis. On the other hand, there are sophisticated tools for web UI construction for JavaScript: https://reactjs.org/.

It would be wise to write a component that traverses Go source code, extract all necessary relations that might be useful for visualization and exports them to a file. Later, other component written in JavaScript imports given file and generates necessary SVG vector images directly in the browser.

Thus, this project will have two parts written in two languages: Go (analysis and information collection) and JavaScript (visualization and UI).

# Communication between Go and JavaScript parts

Collected data will have non-trivial relations. Using common formats like JSON or XML may require a lot of additional work for serialization/deserialization. Better solution would be to store collected data directly to SQL database on Go side and import it directly from database on JavaScript side. This way we can avoid manual serialization/deserialization.

SQLite database engine allows to store database as a single file. There also exist a pure JavaScript implementation of SQLite, which allows to import database directly into browser as a single file: https://sql.js.org/

Thus, resulting tool will have following workflow:

 1. Run Go command providing path to the project that produces .db file
 2. Start JavaScript webserver
 3. Open browser on localhost with port used by webserver
 4. Upload .db file produced at step 1.
 5. See loaded data visualized, navigate the project

# Collected information

The goal of the project is to provide bird's view on the codebase. While working with unfamiliar code developer often navigates code starting either from entry point function or from input/output calls.

Go is a concurrent language. Spawning a goroutine for a subtask is a very common pattern. Almost always goroutines communicate between each other using channels. It would be very useful to collect all locations where new gouroutines are spawned or where receive/send via channel is done. This information would help a lot to understand where in the code IO is done and how it affects the state of the system.

Often IO and channel send/receive is done indirectly, via subfunctions. Therefore it also would be useful to collect dependencies between function calls.

 Thus, collected information must include:

 * Entry point, main function
 * IO calls locations, function calls that read/write from sockets or files
 * goroutine spawn points
 * channel input/output points
 * dependencies between function calls.

In some cases we wouldn't be able to determine which function will be called (dynamic method dispatch). In that case we leave it blank.
