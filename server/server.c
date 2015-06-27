/*
  Server to lookup domain names in hash table
*/

#include <stdio.h>
#include <string.h>  //strlen
#include <stdlib.h>  //malloc
#include <sys/socket.h>
#include <arpa/inet.h>  //inet_addr
#include <unistd.h>   //write
#include <pthread.h>  //for threading, link with lpthread
#include "uthash.h"   // hash table implementation

struct record {
  char domain[255]; /* key */
  in_addr_t ip;
  UT_hash_handle hh;  /* makes structure hashable */
};

struct record *records = NULL;

void add_record(char *domain_name, in_addr_t ip_addr)
{
  struct record *r;

  HASH_FIND_STR(records, domain_name, r); // domain already in hash?
  if(r == NULL)
  {
    r = (struct record*)malloc(sizeof(struct record));
    strcpy(r->domain, domain_name);
    HASH_ADD_STR(records, domain, r); // domain: name of key field
  }
  r->ip = ip_addr;
}

struct record *ip_lookup(char *domain_name)
{
  struct record *r;

  HASH_FIND_STR(records, domain_name, r); // r: output pointer
  return r;
}

void delete_record(struct record *r)
{
  HASH_DEL(records, r);  // r: pointer to delete
  free(r);
}

//the thread function
void *connection_handler(void *);

int main(int argc, char *argv[])
{
  int socket_desc, client_sock, c, *new_sock;
  struct sockaddr_in server, client;
  char client_message[2000];

  // Add initial records to hash (for testing)
  add_record("www.google.com", inet_addr("74.125.224.72"));
  add_record("www.facebook.com", inet_addr("69.63.176.13"));
  add_record("example.com", inet_addr("93.184.216.119"));

  //Create socket
  socket_desc = socket(AF_INET, SOCK_STREAM, 0);
  if (socket_desc == -1)
  {
    perror("Could not create socket");
    return 1;
  }
  puts("Socket created");

  // Prepare the sockaddr_in structure
  server.sin_family = AF_INET;
  server.sin_addr.s_addr = INADDR_ANY;
  server.sin_port = htons( 8888 );

  //Bind
  if( bind(socket_desc, (struct sockaddr *)&server, sizeof(server)) < 0)
  {
    //print error message
    perror("Bind failed");
    return 1;
  }
  puts("Bind done.");

  //Listen
  listen(socket_desc, 3);

  //Accept an incoming connection
  puts("Waiting for incoming connections...");
  c = sizeof(struct sockaddr_in);
  while(client_sock = accept(socket_desc, (struct sockaddr *)&client, (socklen_t*)&c))
  {
    puts("Connection accepted.");

    pthread_t sniffer_thread;
    new_sock = malloc(1);
    *new_sock = client_sock;

    if(pthread_create(&sniffer_thread, NULL, connection_handler, (void*) new_sock) < 0)
    {
      perror("Could not create thread");
      return 1;
    }

    //Now join the thread so that we don't terminate before the thread
    pthread_join(sniffer_thread, NULL);
    puts("Handler assigned");
  }

  if(client_sock < 0)
  {
    perror("Accept failed.");
    return 1;
  }

  return 0;
}

/*
 * This will handle connection for each client
 */
void *connection_handler(void *socket_desc)
{
  //Get the socket descriptor
  int sock = *(int*)socket_desc;
  int read_size;
  char *message, client_message[2000];
  struct record *rec;
  struct in_addr ip_addr;

  //Receive a message from the client
  while( (read_size = recv(sock, client_message, 2000, 0)) > 0)
  {
    message = (char *)malloc(300);
    // Send the message back to the client
    // write(sock, client_message, strlen(client_message));
    printf(client_message);
    rec = ip_lookup(client_message);
    if(rec == NULL)
    {
       message = "Could not find domain name ";
       strcat(message, client_message);
       strcat(message, "!");
       printf(message);
    }
    else
    {
      ip_addr.s_addr = rec->ip;
      message = inet_ntoa(ip_addr);
      printf(message);
    }
    write(sock, message, strlen(message));
    memset(client_message, '\0', sizeof(client_message));
  }

  if(read_size == 0)
  {
    puts("Client disconnected.");
    fflush(stdout);
  }
  else if(read_size == -1)
  {
    perror("recv failed");
    free(socket_desc);
    return (void*)1;
  }

  free(socket_desc);
  free(message);
  free(rec);
  return (void*)0;

}
