import os
import redis
from flask import Flask
from flask import jsonify
import json
app = Flask(__name__)
db=redis.Redis(host='127.0.0.1')


@app.route('/')
def hello_world():
    d = db.keys()
    q = [ rec.decode('ascii') for rec in d ]
    return str(q)

@app.route('/details/<aaa>') 
def foo(aaa):
    d = db.hgetall(aaa)
    q = { y.decode('ascii'): d.get(y).decode('ascii') for y in d.keys() }
    return  q

@app.route('/setname/<name>')
def setname(name):
    #db.set('name',name)
    return 'Name updated.'

if __name__ == '__main__':
    app.run(host= '0.0.0.0', port='5055')
