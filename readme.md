<!-- ABOUT THE PROJECT -->
## About The Project
This is a simple tool I built to parse the english [wikipedia xml dump](https://dumps.wikimedia.org/). It uses goroutines to parse the 105gb(at the time of writing) file and create a map that pairs a pages name with the links it contains to other pages. It's basically an adjacency list in NDJSON format.

<!-- Built WIth -->
## Built With
[Golang](https://go.dev/)

<!-- LICENSE -->
## License
Distributed under the [GPL V3 license](https://www.gnu.org/licenses/gpl-3.0.en.html).