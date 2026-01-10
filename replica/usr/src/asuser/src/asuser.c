#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>

int main(int argc, char* argv[])
{
    if (argc<3)
    {
        fprintf(stderr, "Usage: asuser <target_user> <command> [args...]\n");
        return 1;
    }

    printf("Target user : %s\n", argv[1]);
    printf("Executing   : %s\n", argv[2]);

    execvp(argv[2], &argv[2]);
    perror("execvp");
    return 1;
}
