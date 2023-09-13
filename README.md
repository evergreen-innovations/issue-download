# issue-download

`Issue-download` downloads the issues for a given repo and writes them to a simple text file, along
with any images. This allows issues to be read offline if required.

## Getting started
The tool is written in Go and can be compiled with any currently supported version


```
go build
```

In order for the tool to access the Github respositories a personal access token is requried. Details
of generating these tokens can be found it the Github [documentation](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens).

The enviroment variable `GH_TOKEN` must be present when running the tool.

The tool takes two arguments, the repository owner and the repository

## Running the tool
Once the tool is build it can be run with

```
./issue-download owner repo
``

For example, to download the issues for this repository (assuming you have a personal access token with permissions) would be

```
./issue-download evergreen-innovations issue-download
```
