# Enterprise URL Shortener: AWS Microservice Architecture

A full-stack, highly available URL shortener demonstrating modern cloud system design, deployment pipelines, and containerized orchestration. 

The application features a React/Vite frontend hosted on S3, and a high-performance Go API managed by AWS ECS Fargate, backed by PostgreSQL and Redis.

## 🏗️ System Architecture

<img width="1938" height="1525" alt="Image" src="https://github.com/user-attachments/assets/eb1162bf-5f35-44fd-85bf-919456ecfabb" />

### Key Engineering Patterns
* **Cache-Aside:** The Go API intercepts read requests and checks Redis for the short code before hitting the Postgres database, reducing database read volume.
* **Token Bucket Rate Limiting:** The API utilizes Redis to enforce rate limiting on incoming requests to prevent abuse and DDoS vectors.
* **Base62 Encoding**: The API utilizes Base62 encoding ([0-9a-zA-Z]) to deterministically compress sequential Postgres database IDs into URL-safe alphanumeric short codes. 

## 🚀 Deployment Guide

This project utilizes a bifurcated deployment strategy: **Manual Infrastructure Deployment** combined with **Automated Application Deployment (GitOps)**.

### Prerequisites
* AWS CLI configured locally with administrative permissions.
* Docker desktop running.
* Node.js and `pnpm` installed.

### 1. Deploy Base Infrastructure & Security
First, provision the Elastic Container Registry (ECR) and the GitHub OIDC authentication role.

1. Deploy the base stack:
   ```bash
   aws cloudformation create-stack \
     --stack-name shortener-base-infra \
     --template-body file://infra/base.yaml \
     --capabilities CAPABILITY_NAMED_IAM \
     --parameters ParameterKey=GitHubOrg,ParameterValue=YOUR_USERNAME \
                  ParameterKey=GitHubRepo,ParameterValue=YOUR_REPO_NAME
    ```

2. Retrieve the `GithubRoleArn` from the stack outputs and save it as a Repository Secret in Github named `AWS_ROLE_ARN`

### 2. Push Docker Image to ECR

Because the ECS will need the Backend App docker image in ECR to successfully deploy, this needs to be done before deploying the full infra. This can be done in either of two ways:
- (Preferred) Manually trigger the github action. The image push will succeed, but the entire action will fail because ECS hasn't been deployed yet.
- Manually push the image with your CLI:
    ```bash
    export AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query "Account" --output text)

    docker build -t shortener-api ./backend
    
    docker tag shortener-api $AWS_ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/shortener-backend:latest
    
    docker push $AWS_ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/shortener-backend:latest
    ```

### 3. Deploy Core Infrastructure
1. Deploy the main template:
    ```bash
        aws cloudformation deploy \
        --stack-name shortener-prod-infra \
        --template-file infra/template.yaml \
        --capabilities CAPABILITY_IAM
    ```

2. Once the stack deploy is complete (5-10 minutes), retrieve the `ApiUrl` (the load balancer DNS) from the stack output.

3. Save this as a Repository Variable in Github named `VITE_API_BASE_URL`

### 4. Final Sync

Rerun the frontend action in github to compile the React application and sync to the newly created S3 bucket.

### 5. Teardown
1. Empty the S3 bucket:
    ```bash
    aws s3 rm s3://YOUR-BUCKET-NAME --recursive
    ```

2. Destroy the Core Infrastructure
    ```bash
    aws cloudformation delete-stack --stack-name shortener-prod-infra
    ```
