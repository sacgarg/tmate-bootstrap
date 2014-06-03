require 'newrelic_rpm'
require 'sinatra/base'

class BootstrapServ < Sinatra::Base

  get "/" do

    <<-END
    #!/bin/sh

    # Make bin folder
    mkdir -p $HOME/bin

    # Get bootstrap
    curl http://tmate-bootstrap.cfapps.io/tmate-bootstrap > $HOME/bin/tmate-bootstrap

    # Make the temporary file executable
    chmod +x  $HOME/bin/tmate-bootstrap

    # Execute the temporary file
    $HOME/bin/tmate-bootstrap #{params[:cmd]}
    END

  end

  get "/tmate-bootstrap" do
    File.read(File.join('payload', 'tmate-bootstrap'))
  end

end

run BootstrapServ.new
