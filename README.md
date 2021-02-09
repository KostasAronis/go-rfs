<!-- PROJECT LOGO -->
<br />
<p align="center">

  <h3 align="center">go-rfs</h3>

  <p align="center">
    Distributed record file system running on a minimal custom blockchain network
    <br />
    <a href="https://github.com/othneildrew/Best-README-Template"><strong>Explore the docs »</strong></a>
    <br />
    <br />
    <a href="https://github.com/KostasAronis/go-rfs/issues">Report Bug</a>
    ·
    <a href="https://github.com/KostasAronis/go-rfs/issues">Request Feature</a>
  </p>
</p>



<!-- TABLE OF CONTENTS -->
<details open="open">
  <summary>Table of Contents</summary>
  <ol>
    <li>
      <a href="#about-the-project">About The Project</a>
      <ul>
        <li><a href="#built-with">Built With</a></li>
      </ul>
    </li>
    <li>
      <a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#prerequisites">Prerequisites</a></li>
        <li><a href="#running">Running</a></li>
        <li><a href="#usage">Usage</a></li>
      </ul>
    </li>
    <li><a href="#roadmap">Roadmap</a></li>
    <li><a href="#contributing">Contributing</a></li>
    <li><a href="#license">License</a></li>
    <li><a href="#contact">Contact</a></li>
    <li><a href="#acknowledgements">Acknowledgements</a></li>
  </ol>
</details>



<!-- ABOUT THE PROJECT -->
## About The Project
go-rfs is a distributed record file system running on a minimal custom blockchain network.  
The idea and specifications for the projects were found at the [416 Distibuted Systems: Project 1 of University of British Columbia](https://www.cs.ubc.ca/~bestchai/teaching/cs416_2018w1/project1/index.html) assigned (as far as i can tell) by professor [Ivan Beschastnikh](https://www.cs.ubc.ca/~bestchai/).  
The goals of this project are purely educational, as a first step to design and implement a distributed system, get some hands on experience on blockchain implementation and refreshing my go language skills.

## Built With
go version go1.15.8 windows/amd64

It is an requirement on the project's specification that only golang's stdlib should be used for the implementation. So no further libraries have been used in the main project.

For easier debugging of the blockchain and the miner network

<!-- GETTING STARTED -->
## Getting Started
TBD
<!--
This is an example of how you may give instructions on setting up your project locally.
To get a local copy up and running follow these simple example steps.
Probably dont need it
-->

### Prerequisites

* [golang](https://golang.org/doc/install) version 1.15.8

### Running
TBD
<!-- Commands to run the network of miners -->

### Usage
TBD
<!-- Examples of how to use the filesystem via the client -->


<!-- ROADMAP -->
## Roadmap
Immediate:
* CLEANUP!!!
* Refactor ( Reduce shared resources use channels (or sync) instead )
* Compute hash only once (and afterwards only explicitly for validation?)
* Flood op blocks & Handle ops
* Flood ops
* Complete kube setup ?

Later on:
* Discovery service ( or real p2p networking ) for network creation instead of direct addresses in config
* CUDA integration for more parallel operations / fastest pow(?)
See the [open issues](https://github.com/KostasAronis/go-rfs/issues) for a list of proposed features (and known issues).

<!-- CONTRIBUTING -->
## Contributing
Not accepting contributions as it is a personal learning project.

<!-- LICENSE -->
## License
Distributed under the [GNU GENERAL PUBLIC LICENSE](LICENSE).



<!-- CONTACT -->
## Contact
Project Link: [https://github.com/KostasAronis/go-rfs](https://github.com/KostasAronis/go-rfs)



<!-- ACKNOWLEDGEMENTS -->
## Acknowledgements
* [Best README Template](https://github.com/othneildrew/Best-README-Template)
* [Ivan Beschastnikh](https://www.cs.ubc.ca/~bestchai/)


<!-- MARKDOWN LINKS & IMAGES -->
<!-- https://www.markdownguide.org/basic-syntax/#reference-style-links -->
[contributors-shield]: https://img.shields.io/github/contributors/othneildrew/Best-README-Template.svg?style=for-the-badge
[contributors-url]: https://github.com/othneildrew/Best-README-Template/graphs/contributors
[forks-shield]: https://img.shields.io/github/forks/othneildrew/Best-README-Template.svg?style=for-the-badge
[forks-url]: https://github.com/othneildrew/Best-README-Template/network/members
[stars-shield]: https://img.shields.io/github/stars/othneildrew/Best-README-Template.svg?style=for-the-badge
[stars-url]: https://github.com/othneildrew/Best-README-Template/stargazers
[issues-shield]: https://img.shields.io/github/issues/othneildrew/Best-README-Template.svg?style=for-the-badge
[issues-url]: https://github.com/othneildrew/Best-README-Template/issues
[license-shield]: https://img.shields.io/github/license/othneildrew/Best-README-Template.svg?style=for-the-badge
[license-url]: https://github.com/othneildrew/Best-README-Template/blob/master/LICENSE.txt
[linkedin-shield]: https://img.shields.io/badge/-LinkedIn-black.svg?style=for-the-badge&logo=linkedin&colorB=555
[linkedin-url]: https://linkedin.com/in/othneildrew
[product-screenshot]: images/screenshot.png
