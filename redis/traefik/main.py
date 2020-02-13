import redis
import sys


def test_connection(host):

    try:
        # Connect to redis database
        conn = redis.StrictRedis(host, port=6379)
        if conn is not None:
            response = conn.info()
            role = response["role"]
            print(role)
            return True
        elif conn is None:
            return False
    except Exception as ex:
        print("Exception caught {}".format(ex))


def usage():
    usage = """-------------------------------------------------------------------------
    Mark Day (mark.day@aistemos.com) 01/10/2019
    Tests a redis connection."
    -------------------------------------------------------------------------"""
    print(usage)
    sys.exit(' ')


def main():
    response = test_connection("localhost")
    print(response)


if __name__ == "__main__":
    main()