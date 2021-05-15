import os

import CliParser
import BasicCli
import TerminalUtil


class BirdcCmd(object):
    syntax = """birdc [ <arg> ]"""
    birdcArgsRule = CliParser.StringRule("ARG", "Executable arguments and/or switches", 'args')
    data = {'birdc': 'enter bird shell', '<arg>': birdcArgsRule}

    # noinspection PyMethodMayBeStatic
    def birdc(self, mode, args):
        with TerminalUtil.NoTerminalSettings(mode.session_.terminalCtx_):
            command = "sudo birdc -s /run/bird.ctl"
            if args:
                command += " " + args
            os.system(command)

    def handler(self, mode, args):
        self.birdc(mode, args.get("<arg>", None))


# Register the command by adding a new class
BasicCli.EnableMode.addCommandClass(BirdcCmd)
