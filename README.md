# Pulse - Terraform time estimation service

Use Pulse service - to estimate the time it takes to provision your Terraform template, based on the Terraform plan file.  The Terraform template provisioning can become more predictable and will let the DevOps, SRE and engineers plan their daily activities.

---

## About this repository

This repository contains the following component:
* Pulse REST API server
* Pulse ??? 

The Pulse service integrates with the following:
* Time-estimation data store (fallback db for estimation data - in MongoDB)
* Time-estimation ML model (learnt from raw data, generated using terragrunt test automation)

The API Input includes the following:
* Terraform Plan JSON file, with the following details:
  * Resource name 
  * Region
  * Date & Time
  * ServiceType

The API Output includes the following:
* Estimated total time to provision the resources in the Terraform plan file
* Estimated time taken to provision the individual resources

---

## Environment variables

| Name | Description |  Required |
|------|-------------|-----------|
| <a name="PRIMARY_DB"></a> [PRIMARY\_DB](#PRIMARY\_DB) | Name of Primary DB | yes |
| <a name="FALLBACK_DB"></a> [FALLBACK\_DB](#FALLBACK\_DB) | Name of the Fallback DB | yes |
| <a name="ML_API_ENDPOINT"></a> [ML\_API\_ENDPOINT](#ML\_API\_ENDPOINT) | ML Endpoint where Watson studio generated a model out of historical data  | yes |
| <a name="DB_URL"></a> [DB\_URL](#DB\_URL) | Mongo DB URL | yes |
| <a name="IC_API_KEY"></a> [IC\_API\_KEY](#IC\_API\_KEY) | IBM Cloud API Key | yes |

## How to build and test

* Set you IBM Cloud API key
   ```
   export IC_API_KEY=<your_api_key>
   ```
* Generate the terraform plan output into json format
   ```
   terraform plan --out tfplan.binary`
   terraform show -json tfplan.binary > tfplan.json`
   ```
* Build the server & run the Pulse application
   ```
   make build       // will start the app
   ```
* Build docker image
   ```
   docker build -t pulse .
   ```
