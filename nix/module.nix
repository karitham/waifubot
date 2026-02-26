{ config, lib, ... }:
let
  mkPort =
    defaultPort:
    lib.mkOption {
      type = lib.types.port;
      default = defaultPort;
      description = "Port to listen on";
    };

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

    bot = lib.mkOption {
      type = lib.types.bool;
      default = true;
      description = "Enable waifubot bot";
    };

    botSettings = lib.mkOption {
      type = lib.types.submodule {
        options = {
          port = mkPort 8080;
          logLevel = mkLogLevel;
        };
      };
      default = {
        port = 8080;
        logLevel = "INFO";
      };
      description = "Bot service settings";
    };

    api = lib.mkEnableOption "waifubot api";

    apiSettings = lib.mkOption {
      type = lib.types.submodule {
        options = {
          port = mkPort 3333;
          logLevel = mkLogLevel;
        };
      };
      default = {
        port = 3333;
        logLevel = "INFO";
      };
      description = "API service settings";
    };
  };

  config = lib.mkIf (cfg.enable || cfg.bot) {
    systemd.services.waifubot-bot = lib.mkIf cfg.bot {
      description = "Waifubot Discord Bot";
      wantedBy = [ "multi-user.target" ];
      after = [ "network.target" ];

      serviceConfig = {
        Type = "simple";
        Restart = "always";
        RestartSec = 10;
        EnvironmentFile = cfg.secretsFile;
        ExecStart = "${cfg.package}/bin/bot run";
        WorkingDirectory = cfg.dataDir;
        RuntimeDirectory = "waifubot";
        StateDirectory = "waifubot";
        PrivateTmp = true;
        NoNewPrivileges = true;
        ProtectSystem = "strict";
        ProtectHome = true;
        ReadWritePaths = [ cfg.dataDir ];
        Environment = [
          "LOG_LEVEL=${cfg.botSettings.logLevel}"
          "PORT=${toString cfg.botSettings.port}"
        ];
      };
    };

    systemd.services.waifubot-api = lib.mkIf (cfg.enable || cfg.api) {
      description = "Waifubot API";
      wantedBy = [ "multi-user.target" ];
      after = [ "network.target" ];

      serviceConfig = {
        Type = "simple";
        Restart = "always";
        RestartSec = 10;
        EnvironmentFile = cfg.secretsFile;
        ExecStart = "${cfg.package}/bin/api";
        WorkingDirectory = cfg.dataDir;
        RuntimeDirectory = "waifubot-api";
        StateDirectory = "waifubot-api";
        PrivateTmp = true;
        NoNewPrivileges = true;
        ProtectSystem = "strict";
        ProtectHome = true;
        ReadWritePaths = [ cfg.dataDir ];
        Environment = "PORT=${toString cfg.apiSettings.port}";
      };
    };
  };
}
