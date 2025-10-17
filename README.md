# Journeys cars app tracker 

The Journeys Cars App Tracker service implements a simple API to manage and track the assignment of cars to journeys based on car seat capacity and the number of people traveling in each group.

Cars in the fleet can have 4, 5, or 6 seats. Users request journeys in groups of 1 to 6 people, and members of the same group must ride together. Any group can be assigned to any car with enough empty seats, regardless of the car’s current location. If no suitable car is available, the group will wait until a car becomes free. Once a car is assigned, the group will travel until their drop-off point — groups cannot be swapped to another car to make room for others.

In terms of fairness: groups should be served as quickly as possible while respecting the arrival order whenever feasible. A later-arriving group can only be served before an earlier group if no car can serve the earlier group.

Example: if a group of 6 is waiting and a car with 4 empty seats becomes available, a group of 2 arriving afterward may take those seats. The larger group may have to wait longer, potentially until they leave out of frustration.


## API

The interface provided by the service is a RESTfull API. The operations are as follows.

### GET /status

Indicate the service has started up correctly and is ready to accept requests.

Responses:

* **200 OK** When the service is ready to receive requests.

### PUT /cars

Load the list of available cars in the service and remove all previous data (existing journeys and cars). This method may be called more than once during the life cycle of the service.

**Body** _required_ The list of cars to load.

**Content Type** `application/json`

Sample:

```json
[
  {
    "id": 1,
    "seats": 4
  },
  {
    "id": 2,
    "seats": 6
  }
]
```

Responses:

* **200 OK** When the list is registered correctly.
* **400 Bad Request** When there is a failure in the request format, expected headers, or the payload can't be unmarshalled.

### POST /journey

A group of people requests to perform a journey.

**Body** _required_ The group of people that wants to perform the journey

**Content Type** `application/json`

Sample:

```json
{
  "id": 1,
  "people": 4
}
```

Responses:

* **200 OK** or **202 Accepted** When the group is registered correctly
* **400 Bad Request** When there is a failure in the request format or the payload can't be unmarshalled.

### POST /dropoff

A group of people requests to be dropped off. Whether they traveled or not.

**Body** _required_ A form with the group ID, such that `ID=X`

**Content Type** `application/x-www-form-urlencoded`

Responses:

* **200 OK** or **204 No Content** When the group is unregistered correctly.
* **404 Not Found** When the group is not to be found.
* **400 Bad Request** When there is a failure in the request format or the payload can't be unmarshalled.

### POST /locate

Given a group ID such that `ID=X`, return the car the group is traveling
with, or no car if they are still waiting to be served.

**Body** _required_ A url encoded form with the group ID such that `ID=X`

**Content Type** `application/x-www-form-urlencoded`

**Accept** `application/json`

Responses:

* **200 OK** With the car as the payload when the group is assigned to a car.
* **204 No Content** When the group is waiting to be assigned to a car.
* **404 Not Found** When the group is not to be found.
* **400 Bad Request** When there is a failure in the request format or the payload can't be unmarshalled.








