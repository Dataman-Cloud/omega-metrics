# omega-app
Omega application service

### Building

            cd $GOPATH/src/github.com/Dataman-Cloud
            git clone git@github.com:Dataman-Cloud/omega-app.git
            cd omega-app
            go build

### Configuration 
The omega-app.yaml can be placed in multiple paths: /etc/omega/, ~/.omega, ./ and the last has the highest priority

            mkdir ~/.omega
            cp omega-app.yaml.sample ~/.omega/omega-app.yaml
            # change the settings to your env
            
### Adding new 3rd libs
The project current use [gvt](http://github.com/FiloSottile/gvt) as deps management tool, it's using GO15VENDOREXPERIMENT of go 1.5, it fetchs deps iteratively and save them in ./vendor/ directory

            go get -u github.com/FiloSottile/gvt
            # resolve gvt conflict
            echo 'alias ggvt='$GOPATH/bin/gvt' >> ~/.bashrc
            # gvt will fetch deps iteratively
            ggvt fetch github.com/Someone/something
