import os

def get_queue_addresses():
    # Service names as defined in the docker-compose file
    queue_services = ["queue1", "queue2"]

    # The internal port for the queues as defined in the docker-compose file
    queue_port = 9000

    # Build the addresses
    addresses = [f"{service}:{queue_port}" for service in queue_services]

    return addresses

if __name__ == "__main__":
    addresses = get_queue_addresses()
    for address in addresses:
        print(f"Queue address: {address}")


