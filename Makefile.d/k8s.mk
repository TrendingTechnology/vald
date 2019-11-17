#
# Copyright (C) 2019 Vdaas.org Vald team ( kpango, kou-m, rinx )
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
.PHONY: k8s/vald/deploy
## deploy vald sample cluster to k8s
k8s/vald/deploy:
	kubectl apply -f k8s/agent/ngt
	kubectl apply -f k8s/discoverer/k8s
	kubectl apply -f k8s/external/redis
	kubectl apply -f k8s/meta/redis
	kubectl apply -f k8s/external/mysql
	kubectl apply -f k8s/manager/backup/mysql
	sleep 2
	kubectl apply -f k8s/gateway/vald

.PHONY: k8s/vald/remove
## remove vald sample cluster from k8s
k8s/vald/remove:
	kubectl delete -f k8s/gateway/vald
	kubectl delete -f k8s/manager/backup/mysql
	kubectl delete -f k8s/external/mysql
	kubectl delete -f k8s/meta/redis
	kubectl delete -f k8s/external/redis
	kubectl delete -f k8s/discoverer/k8s
	kubectl delete -f k8s/agent/ngt
