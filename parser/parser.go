package parser

import (
	"fmt"
	"reflect"
	"strings"
)

func GetResourceList(input map[string]interface{}) []string {
	resourceList := make([]string, 0)
	plannedValues := input["planned_values"].(map[string]interface{})
	rootModule := plannedValues["root_module"].(map[string]interface{})
	ResourceList := rootModule["resources"].([]interface{})
	for _, resource := range ResourceList {
		info := resource.(map[string]interface{})
		resourceList = append(resourceList, info["address"].(string))
	}
	return resourceList
}

//Returns provider info like region etc
func GetProviderInfo(input map[string]interface{}) map[string]map[string]interface{} {

	/*******************************
		SAMPLE
		"provider_config": {
	            "ibm": {
	                "name": "ibm",
	                "expressions": {
	                    "region": {
	                        "constant_value": "us-south"
	                    }
	                }
	            }
	        },
		*******************************/

	providerInfoMap := make(map[string]map[string]interface{})
	ArgumentValuesMap := make(map[string]interface{})

	configuration := input["configuration"].(map[string]interface{})
	if configuration["provider_config"] != nil {
		providerConfigMap := configuration["provider_config"].(map[string]interface{})

		for provider, _ := range providerConfigMap {
			if provider == "ibm" {
				ibmConfigMap := providerConfigMap["ibm"].(map[string]interface{})
				providerInfo := ibmConfigMap["expressions"].(map[string]interface{})
				for key, value := range providerInfo {
					if key == "region" {
						configuredRegion := value.(map[string]interface{})
						//iterate over all the argument values
						for k, v := range configuredRegion {
							if k == "constant_value" {
								ArgumentValuesMap[key] = v
							}
						}
					}
				}
				providerInfoMap["ibm"] = ArgumentValuesMap
			}
		}
	}
	return providerInfoMap
}

func GetArgumentListPerResource(input map[string]interface{}) map[string]map[string]interface{} {
	resourceArgumentValuesMap := make(map[string]map[string]interface{})

	configuration := input["configuration"].(map[string]interface{})
	rootModule := configuration["root_module"].(map[string]interface{})
	ResourceList := rootModule["resources"].([]interface{})

	for _, resource := range ResourceList {

		ArgumentValuesMap := make(map[string]interface{})
		resourceInfo := resource.(map[string]interface{})
		// address is combination of type and name,
		// e.g: ibm_resource_instance.activity_tracker  (type.name)
		resourceName := resourceInfo["address"].(string)

		//skip data sources
		if resourceInfo["mode"].(string) == "data" {
			continue
		}
		expression := resourceInfo["expressions"].(map[string]interface{})
		for key, value := range expression {
			if reflect.ValueOf(value).Kind() == reflect.Map {
				/*
					Handling case
					"bucket_name": {
						"constant_value": "mybucket-mine012090"
					},

					"allowed_ip": {
					    "constant_value": [
					        "127.0.0.1",
					    	"192.168.55.102",
					    ]
					},
				*/
				if key == "location" || key == "service" {
					argumentValues := value.(map[string]interface{})
					//iterate over all the argument values
					for k, v := range argumentValues {
						if k == "constant_value" {
							ArgumentValuesMap[key] = v
						}
					}
				}
			}
			/*else if reflect.ValueOf(value).Kind() == reflect.Slice {
				/*
										Handling Caes

										"activity_tracking": [
										{
										    "activity_tracker_crn": {
										        "references": [
										            "ibm_resource_instance.activity_tracker"
										        ]
										    },
										    "read_data_events": {
										        "constant_value": true
										    },
										    "write_data_events": {
										        "constant_value": true
										    }
										}
										],

										 "resources": [
					                            	{
					                                	"service": {
					                                    	"constant_value": "cloud-object-storage"
					                                	}
					                            	}
					                        ],


				list := value.([]interface{})
				//nestedArguments := make(map[string]interface{})
				for _, i := range list {
					argumentValues := i.(map[string]interface{})
					//iterate over all the argument values
					for k, v := range argumentValues {
						if k == "location" || k == "service" {
							arg := v.(map[string]interface{})
							for ref, dep := range arg {
								if ref == "constant_value" {
									if reflect.ValueOf(dep).Kind() == reflect.String {
										//nestedArguments[k] = dep.(string)
										ArgumentValuesMap[k] = dep.(string)
									} else if reflect.ValueOf(dep).Kind() == reflect.Bool {
										//nestedArguments[k] = dep.(bool)
										ArgumentValuesMap[k] = dep.(bool)
									}
								}
							}
						}
					}
				}
				//ArgumentValuesMap[key] = nestedArguments
			}*/
		}
		resourceArgumentValuesMap[resourceName] = ArgumentValuesMap
	}

	//Getting the action ( create / update / delete/ no-op )
	resource_changes := input["resource_changes"].([]interface{})

	/*
			"resource_changes": [
		        {
		            "address": "ibm_cos_bucket.standard-ams03",
		            "mode": "managed",
		            "type": "ibm_cos_bucket",
		            "name": "standard-ams03",
		            "provider_name": "ibm",
		            "change": {
		                "actions": [
		                    "create"
		                ],
	*/

	for _, i := range resource_changes {
		actionValuesMap := make(map[string]interface{})
		resourceInfo := i.(map[string]interface{})
		resourceName := resourceInfo["address"].(string)
		change := resourceInfo["change"].(map[string]interface{})
		for key, value := range change {
			if key == "actions" {
				v := value.([]interface{})
				actionValuesMap[key] = v[0].(string)
			}
		}

		if attributeMap, exist := resourceArgumentValuesMap[resourceName]; exist {
			//fmt.Println("resource found attributeMap is: ", attributeMap)
			//merge action map with region,service map
			for key, value := range actionValuesMap {
				attributeMap[key] = value
			}

			resourceArgumentValuesMap[resourceName] = attributeMap
			//fmt.Println("After update action : ", attributeMap)
		} else {
			fmt.Println("resource does not exists, adding new entry :", resourceName)
		}

	}

	//fmt.Printf("**** resourceArgumentsMap %v", resourceArgumentValuesMap)
	return resourceArgumentValuesMap
}

func GetUpdatedResourceList(input map[string]interface{}) []string {
	updatedResources := make([]string, 0)

	resource_changes := input["resource_changes"].([]interface{})

	for _, i := range resource_changes {
		//actionValuesMap := make(map[string]interface{})
		resourceInfo := i.(map[string]interface{})
		resourceName := resourceInfo["address"].(string)
		change := resourceInfo["change"].(map[string]interface{})
		for key, value := range change {
			if key == "actions" {
				v := value.([]interface{})
				//actionValuesMap[key] = v[0].(string)
				if v[0].(string) == "update" || v[0].(string) == "create" {
					updatedResources = append(updatedResources, resourceName)
				}
			}
		}
	}
	return updatedResources
}

func PrepareResourceDependecyList(input map[string]interface{}) (map[string][]string, []string) {
	resourceDependencyMap := make(map[string][]string, 0)

	configuration := input["configuration"].(map[string]interface{})
	rootModule := configuration["root_module"].(map[string]interface{})
	ResourceList := rootModule["resources"].([]interface{})
	totalDependencyList := make([]string, 0)
	totalResourceList := make([]string, 0)
	//iterate over all resources
	for _, resource := range ResourceList {
		dependencyList := make([]string, 0)

		resourceInfo := resource.(map[string]interface{})
		//resourceName := resourceInfo["type"].(string)
		resourceName := resourceInfo["address"].(string)
		totalResourceList = append(totalResourceList, resourceName)
		expression := resourceInfo["expressions"].(map[string]interface{})

		//iterate over expression block
		for _, value := range expression {
			if reflect.ValueOf(value).Kind() == reflect.Map {
				/* if the value is map, handling the cases like

								"resource_instance_id": {
				                            "references": [
				                                "ibm_resource_instance.cos_instance"
				                            ]
								},

				*/
				argumentValues := value.(map[string]interface{})
				//iterate over all the argument values
				for k, v := range argumentValues {
					if k == "references" {
						val := v.([]interface{})
						//iterate over all the references (dependent resources)
						for _, s := range val {
							str := s.(string)
							// Dont consider variable references like e.g: region = var.location
							if strings.HasPrefix(str, "var.") || strings.HasPrefix(str, "data.") || strings.HasPrefix(str, "each.") {
								continue
							}
							dependencyList = append(dependencyList, str)
						}
					}
				}
			} else if reflect.ValueOf(value).Kind() == reflect.Slice {
				/* if the value is list like e.g

												"activity_tracking": [
								                            {
								                                "activity_tracker_crn": {
								                                    "references": [
								                                        "ibm_resource_instance.activity_tracker"
								                                    ]
								                                },
								                                "read_data_events": {
								                                    "constant_value": true
								                                },
								                                "write_data_events": {
								                                    "constant_value": true
								                                }
								                            }
												],

												 "resources": [
				                            	{
				                                	"service": {
				                                    	"constant_value": "cloud-object-storage"
				                                	}
				                            	}
				                        		],
				*/

				list := value.([]interface{})
				for _, i := range list {
					argumentValues := i.(map[string]interface{})
					//iterate over all the argument values
					for _, v := range argumentValues {
						arg := v.(map[string]interface{})
						for ref, dep := range arg {
							if ref == "references" {
								val := dep.([]interface{})
								for _, s := range val {
									str := s.(string)
									if strings.HasPrefix(str, "var.") || strings.HasPrefix(str, "data.") || strings.HasPrefix(str, "each.") {
										continue
									}
									dependencyList = append(dependencyList, str)
								}
							}
						}
					}
				}
			}

		}
		//fmt.Println("+++++++++ ResourceName ", resourceName)
		//fmt.Println("+++++++++ dependencyList", dependencyList)
		resourceDependencyMap[resourceName] = dependencyList
		totalDependencyList = append(totalDependencyList, dependencyList...)
	}
	//Get the starting nodes by subtracting dependentnodes from all ndoes
	startingNodes := difference(totalResourceList, totalDependencyList)

	//fmt.Printf("****** starting nodes %v", startingNodes)
	for _, sn := range startingNodes {
		fmt.Println(sn)
	}

	return resourceDependencyMap, startingNodes
}

func SplitString(str string) ([]string, error) {
	if strings.Contains(str, ".") {
		resources := strings.Split(str, ".")
		return resources, nil
	}
	return []string{}, fmt.Errorf("The given id %s does not contain . please check documentation on how to provider id during import command", str)
}

func difference(slice1 []string, slice2 []string) []string {
	diffStr := []string{}
	m := map[string]int{}

	for _, s1Val := range slice1 {
		m[s1Val] = 1
	}
	for _, s2Val := range slice2 {
		m[s2Val] = m[s2Val] + 1
	}

	for mKey, mVal := range m {
		if mVal == 1 {
			diffStr = append(diffStr, mKey)
		}
	}

	return diffStr
}

/*
func PrepareResourceDependecyList(input map[string]interface{}) map[string][]string {
	resourceArgumentsMap := make(map[string][]string, 0)

	configuration := input["configuration"].(map[string]interface{})
	rootModule := configuration["root_module"].(map[string]interface{})
	ResourceList := rootModule["resources"].([]interface{})

	//iterate over all resources
	for _, resource := range ResourceList {
		dependencyList := make([]string, 0)

		resourceInfo := resource.(map[string]interface{})
		//resourceName := resourceInfo["type"].(string)
		resourceName := resourceInfo["address"].(string)
		expression := resourceInfo["expressions"].(map[string]interface{})

		//iterate over expression block
		for _, value := range expression {
			if reflect.ValueOf(value).Kind() == reflect.Map {
				argumentValues := value.(map[string]interface{})
				//iterate over all the argument values
				for k, v := range argumentValues {
					if k == "references" {
						val := v.([]interface{})
						//iterate over all the references (dependent resources)
						for _, s := range val {
							dependencyList = append(dependencyList, s.(string))
						}
					}
				}
			} else if reflect.ValueOf(value).Kind() == reflect.Slice {
				list := value.([]interface{})
				for _, i := range list {
					argumentValues := i.(map[string]interface{})
					//iterate over all the argument values
					for _, v := range argumentValues {
						arg := v.(map[string]interface{})
						for ref, dep := range arg {
							if ref == "references" {
								val := dep.([]interface{})
								for _, s := range val {
									dependencyList = append(dependencyList, s.(string))
								}
							}
						}
					}
				}
			}

		}
		resourceArgumentsMap[resourceName] = dependencyList
	}
	return resourceArgumentsMap
}
*/
