{ config, lib, ... }:
let
  mkLogLevel = lib.mkOption {
    type = lib.types.enum [
      "DEBUG"
      "INFO"
      "WARN"
      "ERROR"
    ];
    default = "INFO";
    description = "Log level";
  };

  cfg = config.services.waifubot;
in
{
  options.services.waifubot = {
    enable = lib.mkEnableOption "waifubot bot and API";

    package = lib.mkOption {
      type = lib.types.package;
      description = "Waifubot package to use";
    };

    secretsFile = lib.mkOption {
      type = lib.types.path;
      description = "Path to secrets environment file";
      example = "/etc/waifubot/secrets.env";
    };

    dataDir = lib.mkOption {
      type = lib.types.path;
      default = "/var/lib/waifubot";
      description = "Data directory for waifubot";
    };

    settings = lib.mkOption {
      type = lib.types.submodule {
        options = {
          port = lib.mkOption {
            type = lib.types.port;
            default = 8080;
            description = "Port to listen on";
          };
          logLevel = mkLogLevel;
          enableApi = lib.mkOption {
            type = lib.types.bool;
            default = true;
            description = "Enable REST API server";
          };
        };
      };
      default = {
        port = 8080;
        logLevel = "INFO";
        enableApi = true;
      };
      description = "Waifubot service settings";
    };
  };

  config = lib.mkIf cfg.enable {
    systemd.services.waifubot = {
      description = "Waifubot Discord Bot and API";
      wantedBy = [ "multi-user.target" ];
      after = [ "network.target" ];

      serviceConfig = {
        Type = "simple";
        Restart = "always";
        RestartSec = 10;
        EnvironmentFile = cfg.secretsFile;
        ExecStart = "${cfg.package}/bin/waifubot run ${
          lib.optionalString (!cfg.settings.enableApi) "--no-api"
        }";
        WorkingDirectory = cfg.dataDir;
        RuntimeDirectory = "waifubot";
        StateDirectory = "waifubot";
        PrivateTmp = true;
        NoNewPrivileges = true;
        ProtectSystem = "strict";
        ProtectHome = true;
        ReadWritePaths = [ cfg.dataDir ];
        Environment = [
          "LOG_LEVEL=${cfg.settings.logLevel}"
          "PORT=${toString cfg.settings.port}"
        ];
      };
    };
  };
}
