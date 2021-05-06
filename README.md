# watchCpuUse

This utility will monitor the cumulative CPU use of a PID and all
child processes under that PID. Specifically this is all the processes
with Parent PID equal to the monitored PID and the children of those
children recursively. When used against the parent process of a
container this gives the CPU use of all processes in that container.

## Usage

```
Usage of ./watchCpuUse [opts] <pid|string>
  -count int
        (optional) Number of sample intervals to monitor. 0 runs forever (default 1)
  -interval int
        (optional) sample interval in seconds (default 10)
  -verbose
        (optional) Verbose output
```

When the argument is a string you will be presented with a list of all
top level processes containing that string in the commandline. The
monitoring will begin when one item is selected. If the string matches
only one process monitoring will be started immediately.

```
[core@cnfocto2 ~]$ ./watchCpuUse -interval 5 -count 3 process-export
  [ 0]    374098 watchCpuUse   ./watchCpuUse -interval 5 -count 3 process-export
  [ 1]    783046 conmon   ...36db28f9f7a59323d9c6f94b7b8df4 --exit-dir /var/run/crio/exits -l /var/log/pods/openshift-monitoring_process-exporter-gg2jc_579d104d-f4f6-459c-945a-35c0d090a332/bd7cc9cbc46f3fda7b5a71c57696587ad836db28...
  [ 2]    783649 conmon   ...ccac32f6ecf6b7a6c2a960fece06a7 --exit-dir /var/run/crio/exits -l /var/log/pods/openshift-monitoring_process-exporter-gg2jc_579d104d-f4f6-459c-945a-35c0d090a332/process-exporter/0.log --log-level info ...
  [ 3]    783661 process-exporte   /bin/process-exporter -config.path /config/config.yml
Enter index of process to monitor: 1
Monitoring pid 783046 "conmon   .../usr/libexec/crio/conmon -b /var/run/containers/storage/overlay-containers/bd7cc9cbc46f3fda7b5a71c57..."
Total 0.000 sec (0.000 %cpu) : 0.000 user, 0.000 sys ( 0.000 % user, 0.000 % sys ) @ 06 May 21 22:50 +0000
Total 0.000 sec (0.000 %cpu) : 0.000 user, 0.000 sys ( 0.000 % user, 0.000 % sys ) @ 06 May 21 22:50 +0000
Total 0.000 sec (0.000 %cpu) : 0.000 user, 0.000 sys ( 0.000 % user, 0.000 % sys ) @ 06 May 21 22:50 +0000
```
