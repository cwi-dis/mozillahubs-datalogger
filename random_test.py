import json
import random
import requests
import sys



def make_request(host):
    data = []

    for _ in range(random.randint(1, 400)):
        data.append([
            "".join(random.choices("abcdefghijklmnopqrstuvwxzy", k=10)),
            random.random(),
            random.random()
        ])

    res = requests.post(
        host,
        data=json.dumps({
            "info": [
                "huh", random.randint(1, 100), random.randint(1, 100)
            ],
            "data": data
        })
    )

    return res.ok


def main():
    if len(sys.argv) < 2:
        print("USAGE:", sys.argv[0], "endpoint")
        sys.exit(1)

    repeat = 3000

    for i in range(repeat):
        print(f"Sending request {i+1}/{repeat}...", end=" ")

        ok = make_request(sys.argv[1])
        print("OK" if ok else "ERR")


if __name__ == "__main__":
    main()
