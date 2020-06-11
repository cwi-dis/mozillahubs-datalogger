import json
import requests
import sys

from random import choices, randint, random


def make_request(host):
    data = []

    for _ in range(randint(1, 400)):
        data.append([
            "".join(choices("abcdefghijklmnopqrstuvwxzy", k=10)),
            random(),
            random()
        ])

    res = requests.post(
        host,
        data=json.dumps({
            "info": ["huh", randint(1, 100), randint(1, 100)],
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
