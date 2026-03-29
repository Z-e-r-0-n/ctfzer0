#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <pwd.h>
#include <string.h>
#include <errno.h>
#include <sys/stat.h>
#include <fcntl.h>

#define TOKEN_DIR "/etc/asusr"
#define TOKEN_LEN 32
int main(){
uid_t uid = getuid();
struct passwd *pw = getpwuid(uid);
if (uid==0)
{
	fprintf(stderr, "root token generation is not allowed\n");
	return 1;
}
if (!pw) {
    perror("getpwuid");
    return 1;
}

const char *username = pw->pw_name;

if (mkdir(TOKEN_DIR, 0755) == -1) {
    if (errno != EEXIST) {
        perror("mkdir /etc/asusr");
        return 1;
    }
}

char token_path[256];
snprintf(token_path, sizeof(token_path),
         "%s/%s.token", TOKEN_DIR, username);


if (access(token_path, F_OK) == 0) {
    printf("Token already exists for %s\n", username);
    printf("1) Rotate token\n");
    printf("2) Remove token\n");
    printf("3) Cancel\n");
    printf("> ");

    int choice = getchar();

    if (choice == '2') {
        unlink(token_path);
        printf("Token removed\n");
        return 0;
    }

    if (choice != '1') {
        printf("Cancelled\n");
        return 0;
    }
}

char token[TOKEN_LEN + 1];
int fd = open("/dev/urandom", O_RDONLY);

if (fd < 0) {
    perror("open /dev/urandom");
    return 1;
}

unsigned char raw[TOKEN_LEN];
read(fd, raw, TOKEN_LEN);
close(fd);

for (int i = 0; i < TOKEN_LEN; i++) {
    token[i] = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
               [raw[i] % 62];
}
token[TOKEN_LEN] = '\0';


int out = open(token_path,
               O_WRONLY | O_CREAT | O_TRUNC,
               0640);

if (out < 0) {
    perror("open token file");
    return 1;
}

write(out, token, strlen(token));
write(out, "\n", 1);
close(out);


chown(token_path, uid, 0);   
chmod(token_path, 0640);


printf("Your asuser token:\n%s\n", token);
printf("Store it securely. It will not be shown again.\n");
}
