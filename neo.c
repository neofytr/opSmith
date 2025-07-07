#include "buildsysdep/neobuild.h"
#include <stdio.h>
#include <string.h>
#include <stdlib.h>

#define WINDOWS "Windows/"
#define LINUX "Linux/"

#define BIN "./bin/"
#define CMD "./cmd/"
#define MASTER "master/"
#define SLAVE "slave/"

void clean_build_artifacts()
{
    remove(BIN LINUX "master");
    remove(BIN LINUX "slave");
    remove(BIN WINDOWS "master.exe");
    remove(BIN WINDOWS "slave.exe");
    remove("slave.tmp");
}

int main(int argc, char **argv)
{
    neocmd_t *linux_master, *linux_slave, *windows_master, *windows_slave;
    bool run = false;
    neorebuild("neo.c", argv, &argc);
    clean_build_artifacts();

    for (int i = 1; i < argc; ++i)
    {
        if (!strcmp(argv[i], "clean"))
        {
            printf("Cleaned build artifacts\n");
            return EXIT_SUCCESS;
        }

        if (!strcmp(argv[i], "run"))
        {
            run = true;
        }
    }

    linux_master = neocmd_create(BASH);
    neocmd_append(linux_master, "CGO_ENABLED=0");
    neocmd_append(linux_master, "GOOS=linux");
    neocmd_append(linux_master, "GOARCH=amd64");
    neocmd_append(linux_master, "go build -o " BIN LINUX "master " CMD MASTER "main.go");
    neocmd_run_sync(linux_master, NULL, NULL, false);

    linux_slave = neocmd_create(BASH);
    neocmd_append(linux_slave, "CGO_ENABLED=0");
    neocmd_append(linux_slave, "GOOS=linux");
    neocmd_append(linux_slave, "GOARCH=amd64");
    neocmd_append(linux_slave, "go build -o " BIN LINUX "slave " CMD SLAVE "main.go");
    neocmd_run_sync(linux_slave, NULL, NULL, false);

    windows_master = neocmd_create(BASH);
    neocmd_append(windows_master, "CGO_ENABLED=0");
    neocmd_append(windows_master, "GOOS=windows");
    neocmd_append(windows_master, "GOARCH=amd64");
    neocmd_append(windows_master, "go build -o " BIN WINDOWS "master.exe " CMD MASTER "main.go");
    neocmd_run_sync(windows_master, NULL, NULL, false);

    windows_slave = neocmd_create(BASH);
    neocmd_append(windows_slave, "CGO_ENABLED=0");
    neocmd_append(windows_slave, "GOOS=windows");
    neocmd_append(windows_slave, "GOARCH=amd64");
    neocmd_append(windows_slave, "go build -o " BIN WINDOWS "slave.exe " CMD SLAVE "main.go");
    neocmd_run_sync(windows_slave, NULL, NULL, false);

    neocmd_delete(linux_master);
    neocmd_delete(linux_slave);
    neocmd_delete(windows_master);
    neocmd_delete(windows_slave);

    if (run)
    {
        neocmd_t *run_slave = neocmd_create(BASH);
        neocmd_append(run_slave, BIN LINUX "slave --port 6969 > slave.tmp &"); // run in background
        neocmd_run_sync(run_slave, NULL, NULL, false);
        neocmd_delete(run_slave);

        neocmd_t *run_master = neocmd_create(BASH);
        neocmd_append(run_master, BIN LINUX "master --client localhost 6969 --run-from-file prompt.txt");
        neocmd_run_sync(run_master, NULL, NULL, false);
        neocmd_delete(run_master);
    }
    return EXIT_SUCCESS;
}
