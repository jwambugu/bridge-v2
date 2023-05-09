# [WIP] - Bridge

API gateway for making bulk payments and bulk notification built using go and gRPC

## Running on Docker [Requires docker]

To run the project on docker, run the following command:

```bash
  make compose-up
```

## Running Tests

To run tests, run the following command:

```bash
  make test
```


### Tasks
- [] Remove categories - changes project from ecommerce 
- [] Add mpesa integration 
- [] Add contacts management
- [] Add queue driver - redis/kafka/rabbitmq?
- [] Add bulk payments
- [] Add firebase notifications support
- [] Add bulk SMS support - africa talking/twilio? 