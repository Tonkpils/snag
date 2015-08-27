
release:
	docker build --tag snag-pre .
	docker run --rm snag-pre bash -c " \
		mv /go/bin/snag snag ;\
		tar -cO snag" >> $@/snag
	docker rmi snag-pre
	docker build --tag snag $@
	rm $@/snag

.PHONY: release