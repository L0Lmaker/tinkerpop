// /*
//  *  Licensed to the Apache Software Foundation (ASF) under one
//  *  or more contributor license agreements.  See the NOTICE file
//  *  distributed with this work for additional information
//  *  regarding copyright ownership.  The ASF licenses this file
//  *  to you under the Apache License, Version 2.0 (the
//  *  "License"); you may not use this file except in compliance
//  *  with the License.  You may obtain a copy of the License at
//  *
//  *  http://www.apache.org/licenses/LICENSE-2.0
//  *
//  *  Unless required by applicable law or agreed to in writing,
//  *  software distributed under the License is distributed on an
//  *  "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
//  *  KIND, either express or implied.  See the License for the
//  *  specific language governing permissions and limitations
//  *  under the License.
//  */
//
// /**
//  * @author Jorge Bay Gondra
//  */
// 'use strict';
//
// const assert = require('assert');
// const Bytecode = require('../../lib/process/bytecode');
// const graphModule = require('../../lib/structure/graph');
// const helper = require('../helper');
// const {traversal} = require("../../lib/process/anonymous-traversal");
// const graphTraversalModule = require("../../lib/process/graph-traversal");
// const __ = graphTraversalModule.statics;
//
// const cache = {};
//
// let connection;
//
// function getVertices(connection) {
//     const g = traversal().withRemote(connection);
//     return g.V().group().by('name').by(__.tail()).next().then(it => it.value);
// }
//
// function getEdges(connection) {
//     const g = traversal().withRemote(connection);
//     return g.E().group()
//         .by(__.project("o", "l", "i").by(__.outV().values("name")).by(__.label()).by(__.inV().values("name")))
//         .by(__.tail())
//         .next()
//         .then(it => {
//             const edges = {};
//             it.value.forEach((v, k) => {
//                 edges[getEdgeKey(k)] = v;
//             });
//             return edges;
//         });
// }
//
// function getEdgeKey(key) {
//     // key is a map
//     return key.get('o') + "-" + key.get('l') + "->" + key.get('i');
// }
//
// describe('DriverRemoteConnection', function () {
//     before(function () {
//         connection = helper.getConnection('gmodern');
//         return connection.open()
//             .then(() => Promise.all([getVertices(connection), getEdges(connection)]))
//             .then(values => {
//                 cache["modern"] = {
//                     connection: connection,
//                     vertices: values[0],
//                     edges: values[1]
//                 };
//             });
//     });
//     after(function () {
//         return connection.close();
//     });
//
//     describe('#submit()', function () {
//         it('should send the request and parse the response', function () {
//             return connection.submit(new Bytecode().addStep('V', []).addStep('tail', []))
//                 .then(function (response) {
//                     assert.ok(response);
//                     assert.ok(response.traversers);
//                     assert.strictEqual(response.traversers.length, 1);
//                     assert.ok(response.traversers[0].object instanceof graphModule.Vertex);
//                 });
//         });
//         it('should send the request with syntax error and parse the response error', function () {
//             return connection.submit(new Bytecode().addStep('SYNTAX_ERROR'))
//                 .catch(function (err) {
//                     assert.ok(err);
//                     assert.ok(err.message.indexOf('599') > 0);
//                     assert.ok(err.statusCode === 599);
//                     assert.ok(err.statusMessage === 'Could not locate method: GraphTraversalSource.SYNTAX_ERROR()');
//                     assert.ok(err.statusAttributes);
//                     assert.ok(err.statusAttributes.has('exceptions'));
//                     assert.ok(err.statusAttributes.has('stackTrace'));
//                 });
//         });
//         it('test ConnectedComponent', function () {
//             // return connection.submit('g.withStrategies(new VertexProgramStrategy({graphComputer:"org.apache.tinkerpop.gremlin.process.computer.GraphComputer"})).V().connectedComponent().has("gremlin.connectedComponentVertexProgram.component")')
//             // g.withComputer().V().connectedComponent().has("gremlin.connectedComponentVertexProgram.component")
//             // .addStep('withComputer',[])
//             // .addStep('V',[])
//             // .addStep('toList',[])
//             // .addStep('tail',[])
//             // .addStep('connectedComponent',[])
//             // .addStep('has',["gremlin.connectedComponentVertexProgram.component"]))
//             return connection.submit(new Bytecode()
//                 .addStep('withComputer',[])
//                 .addStep('V',[])
//                 .addStep('tail',[]))
//                 .then(function (response) {
//                     assert.ok(response);
//                     assert.ok(response.traversers);
//                     assert.strictEqual(response.traversers.length, 1);
//                     assert.ok(response.traversers[0].object instanceof graphModule.Vertex);
//                 });
//         });
//         it('test ConnectedComponent', function () {
//             // return connection.submit('g.withStrategies(new VertexProgramStrategy({graphComputer:"org.apache.tinkerpop.gremlin.process.computer.GraphComputer"})).V().connectedComponent().has("gremlin.connectedComponentVertexProgram.component")')
//             // g.withComputer().V().connectedComponent().has("gremlin.connectedComponentVertexProgram.component")
//             // .addStep('withComputer',[])
//             // .addStep('V',[])
//             // .addStep('toList',[])
//             // .addStep('tail',[])
//             // .addStep('connectedComponent',[])
//             // .addStep('has',["gremlin.connectedComponentVertexProgram.component"]))
//             return connection.submit('g.withStrategies(new VertexProgramStrategy({graphComputer:"org.apache.tinkerpop.gremlin.process.computer.GraphComputer"})).V().connectedComponent().has("gremlin.connectedComponentVertexProgram.component")');
//         });
//     });
// });