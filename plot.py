import matplotlib.pyplot as plt

nodes = []
init_discovery = []
init_variances = []
additional_discovery = []
additional_variances = []

with open("results") as file:
    count = 0
    for line in file:
        if count == 0:
            nodes.append(int(line.split(" ")[0]))
        if count == 1:
            init_discovery.append(float(line.split(" ")[3].strip("s,")))

        if count == 2:
            additional_discovery.append(float(line.split(" ")[7].strip("s,")))
        count = (count + 1) % 3

print(nodes)
print(init_discovery)
print(additional_discovery)

plt.plot(nodes, init_discovery)
plt.ylabel("seconds")
plt.xlabel("nodes")
#plt.errorbar(nodes, init_discovery, range(10))
plt.savefig("init.svg")
plt.close()

plt.plot(nodes, additional_discovery)
plt.ylabel("seconds")
plt.xlabel("nodes")
plt.savefig("additional.svg")
