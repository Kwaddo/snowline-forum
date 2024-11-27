# Forum Project

Welcome to the Forum Project! This project is designed to create a simple and efficient forum for users to discuss various topics.
The project was created by Ahmed Aburowais, Salman Abboudi, Moataz Ibrahim and Ali Khalaf.

## Features

- User registration and authentication
- Create, read, update, and delete posts
- Comment on posts
- Like and dislike posts
- User profiles

## Installation

1. Clone the repository:
    ```bash
    git clone https://learn.reboot01.com/git/ak1/forum.git
    ```
2. Navigate to the project directory:
    ```bash
    cd forum
    ```
3. Run the code:
    ```bash
    go run ./cmd/.
    ```


# To build the docker image using the dockerfile run this command:

    - docker image build -f dockerfile -t forum-image .

# To Build the docker container using the docker file run this command:

    - docker container run -p 3333:3333 --detach --name forum-container forum-image

## Usage

1. Open your browser and navigate to `http://localhost:3333`.
2. Register a new account or log in with an existing account.
3. Start creating and interacting with posts!

# License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
