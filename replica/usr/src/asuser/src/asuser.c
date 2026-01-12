#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <pwd.h>
#include <string.h>
#include <errno.h>


#define TOKEN_DIR "etc/asusr"
#define TOKEN_MAX 128
int read_token(const char *user,char *buf,size_t buflen)
{
	char path[256];
	snprintf(path,sizeof(path),"%s/%s.token", TOKEN_DIR, user);
	
	FILE *f =fopen(path, "r");
	if (!f) return -1;
	
	if (!fgets(buf, buflen,f)) 
{
	fclose(f);
	return -1;
}
	buf[strcspn(buf,  "\n")] =0;
	fclose(f);
	return 0;
}
int main(int argc, char* argv[])
{
	char input[TOKEN_MAX];
	char stored[TOKEN_MAX];
    if (argc<3)
    {
        fprintf(stderr, "Usage: asuser <target_user> <command> [args...]\n");
        return 1;
    }
    uid_t caller_uid= getuid();
    struct passwd *caller_pw = getpwuid(caller_uid);
    if (!caller_pw) {
	perror("getpwuid");
	return 1;
}
printf("caller user: %s\n",caller_pw->pw_name);

struct passwd *target_pw = getpwnam(argv[1]);
if (!target_pw){
	fprintf(stderr, "unknown target user:%s\n",argv[1]);
	return 1;
}
	printf("Traget_UID : %d\n",target_pw->pw_uid);
	printf("Target_GID : %d\n",target_pw->pw_gid);
	
if (read_token(target_pw->pw_name, stored, sizeof(stored)) != 0){
	fprintf(stderr , "no token found for the user\n");
	return 1;
}
	printf("asuser token:");
	fflush(stdout);

if (fgets(input, sizeof(input),stdin)){
	fprintf(stderr, "failed to read token \n");
	return 1;
}
input [strcspn(input, "\n")] = 0;
if (strcmp(input,stored) !=0 ) {
	fprintf(stderr, "authentication failed ");
	return 1;
}
		
   printf("Target user : %s\n", argv[1]);
    printf("Executing   : %s\n", argv[2]);

    execvp(argv[2], &argv[2]);
    perror("execvp");
    return 1;
}
