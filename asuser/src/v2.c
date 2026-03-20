#define _GNU_SOURCE
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <pwd.h>
#include <string.h>
#include <errno.h>
#include <time.h>
#include <sys/wait.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <grp.h>

#define TOKEN_DIR "/etc/asusr"
#define LOG_DIR   "/var/log/asuser"
#define TOKEN_MAX 128

/* ---------- TOKEN ---------- */

int read_token(const char *user, char *buf, size_t buflen) {
    char path[256];
    snprintf(path, sizeof(path), "%s/%s.token", TOKEN_DIR, user);

    FILE *f = fopen(path, "r");
    if (!f) return -1;

    if (!fgets(buf, buflen, f)) {
        fclose(f);
        return -1;
    }

    buf[strcspn(buf, "\n")] = 0;
    fclose(f);
    return 0;
}

/* ---------- LOGGING ---------- */

void log_action(const char *caller,
                const char *target,
                char *const argv[],
                int status)
{
    char path[256];
    snprintf(path, sizeof(path), "%s/%s.log", LOG_DIR, target);

    struct passwd *tpw = getpwnam(target);
    if (!tpw) return;

    int fd = open(path, O_WRONLY | O_CREAT | O_APPEND, 0644);
    if (fd < 0) return;

    /* ensure ownership is target:root */
    fchown(fd, tpw->pw_uid, 0);

    time_t now = time(NULL);
    struct tm tm;
    localtime_r(&now, &tm);

    char ts[64];
    strftime(ts, sizeof(ts), "%F %T", &tm);

    dprintf(fd, "[%s] %s : ", ts, caller);
    for (int i = 2; argv[i]; i++)
        dprintf(fd, "%s ", argv[i]);

    if (WIFEXITED(status))
        dprintf(fd, "exit=%d\n", WEXITSTATUS(status));
    else
        dprintf(fd, "killed\n");

    close(fd);
}

/* ---------- MAIN ---------- */

int main(int argc, char *argv[])
{
    if (argc < 3) {
        fprintf(stderr, "Usage: asuser <target> <command> [args...]\n");
        return 1;
    }

    /* real user must NOT be root */
    if (getuid() == 0) {
        fprintf(stderr, "Refusing to run as root\n");
        return 1;
    }

    struct passwd *caller_pw = getpwuid(getuid());
    struct passwd *target_pw = getpwnam(argv[1]);

    if (!caller_pw || !target_pw) {
        fprintf(stderr, "User lookup failed\n");
        return 1;
    }

    /* token check */
    char stored[TOKEN_MAX], input[TOKEN_MAX];

    if (read_token(target_pw->pw_name, stored, sizeof(stored)) != 0) {
        fprintf(stderr, "No token for %s\n", target_pw->pw_name);
        return 1;
    }

    printf("asuser token: ");
    fflush(stdout);

    if (!fgets(input, sizeof(input), stdin)) {
        fprintf(stderr, "Token read failed\n");
        return 1;
    }

    input[strcspn(input, "\n")] = 0;

    if (strcmp(input, stored) != 0) {
        fprintf(stderr, "Authentication failed\n");
        return 1;
    }

    /* ---------- EXEC ---------- */

    pid_t pid = fork();
    if (pid < 0) {
        perror("fork");
        return 1;
    }

    if (pid == 0) {
        /* child */

        setgroups(0, NULL);
        if (setgid(target_pw->pw_gid) != 0) _exit(1);
        if (seteuid(target_pw->pw_uid) != 0) _exit(1);

        execvp(argv[2], &argv[2]);
        _exit(127);
    }

    int status;
    waitpid(pid, &status, 0);

    log_action(caller_pw->pw_name,
               target_pw->pw_name,
               argv,
               status);

    return 0;
}
