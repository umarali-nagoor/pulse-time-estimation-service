import os
import sys
import requests
 
# total arguments
n = len(sys.argv)

if sys.argv[1] == "deploy":
    os.system('terraform init > /dev/null')
    os.system('terraform plan -out=_planned.tfplan > /dev/null')
    os.system('terraform show -no-color -json _planned.tfplan > _output.json')
    with open('_output.json', 'r') as file:
        data = file.read()
    url = 'http://localhost:8080/api/v1/predictor'
    response = requests.post(url, data=data)
    print("-------------------------------")
    json_resp = response.json()
    job_id = json_resp['jobID']
    yes = input("Do you want to get time estimation details?(yes, why not/nah, I'm good)\n")
    if yes == "yes, why not" or yes == "yes" or yes == "y":
         url = 'http://localhost:8080/api/v1/predictor/' + str(job_id)
         response = requests.get(url)
         print("Time estimation response\n-------------------------------")
         print(response.json())
    print("-------------------------------\napplying terraform things")

elif sys.argv[1] == "get-time":
    os.system('terraform init > /dev/null')
    os.system('terraform plan -out=_planned.tfplan > /dev/null')
    os.system('terraform show -no-color -json _planned.tfplan > _output.json')
    with open('_output.json', 'r') as file:
        data = file.read()
    url = 'http://localhost:8080/api/v1/predictnow'
    response = requests.get(url, data=data)
    print("Time estimation response\n-------------------------------")
    print(response.json())

elif sys.argv[1] == "get-time-and-alts":
    os.system('terraform init > /dev/null')
    os.system('terraform plan -out=_planned.tfplan > /dev/null')
    os.system('terraform show -no-color -json _planned.tfplan > _output.json')
    with open('_output.json', 'r') as file:
        data = file.read()
    url = 'http://localhost:8080/api/v1/predictandgetsuggestion'
    response = requests.get(url, data=data)
    print("Time estimation response with suggestion\n-------------------------------")
    print(response.json())

