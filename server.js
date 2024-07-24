const net = require('net');

// Create a server object
const args = process.argv.slice(2); // This slices off the first two elements
const serverName = args[0]; 
const server = net.createServer((socket) => {
  console.log('New connection established');

  // Handle incoming messages from clients.
  socket.on('data', (data) => {
    console.log('Data received from client: ' + data.toString());
    // Send a customized response back to the client
    const message = data.toString().trim();
    const response = `Your request handled by server No. ${serverName}. he received this: "${message}"`;
    socket.write(response);
  });

  // Handle client disconnection
  socket.on('end', () => {
    console.log('Client disconnected');
  });

  // Handle errors
  socket.on('error', (err) => {
    console.error('Server error: ' + err.message);
  });
});

server.listen(1236, () => {
  console.log(`Server ${serverName} listening on port 1236`);
});
