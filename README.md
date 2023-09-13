# issue-download

`Issue-download` downloads the issues for a given repo and writes them to a simple text file, along
with any images. This allows issues to be read offline if required.

## Getting started
The tool is written in Go and can be compiled with any currently supported version


```
go build
```

In order for the tool to access the Github respositories a personal access token is requried. Details
of generating these tokens can be found it the Github [documentation](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens). You will need a "Classic" PAT.

The enviroment variable `GH_TOKEN` must be present when running the tool.

The tool takes two arguments, the repository owner and the repository

## Running the tool
Once the tool is build it can be run with

```
./issue-download owner repo
```

For example, to download the issues for this repository (assuming you have a personal access token with permissions) would be

```
./issue-download evergreen-innovations issue-download
```

## Output
A directory will be created wherever the issue-download command was run. The directory structure is

```
- owner
  - repo
    * issue_1.txt
    * issue_2.txt
    * ...
    * issue_N.txt
    - assets
        - GROUP_ID_OF_UPLOADED_ASSET
            * UUID_OF_ASSET1.png
            * UUID_OF_ASSET2.png
```

The `assets` directory contains any images that have been uploaded. The sub-directory structure and file naming follows the path of the uploaded image to avoid any ambiguity.

At present, only png images will have the file extension appended. Other types of images will be downloaded and stored but the application to open them will need to be manually selected.
